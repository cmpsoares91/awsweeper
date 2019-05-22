package wipe

import (
	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/iflix/awsweeper/pkg/config"
	"github.com/iflix/awsweeper/pkg/filters"
	"github.com/sirupsen/logrus"
)

type Wiper struct {
	Config *config.Config
}

func (c *Wiper) Run() (aws.IRegionResourceTypeResources, []error, error) {
	var warnings []error
	var resourcesToWipe aws.IRegionResourceTypeResources = make(aws.IRegionResourceTypeResources)

	logrus.WithField("DryMode", c.Config.Options.DryRun).Info()
	for _, region := range c.Config.Options.Regions {
		logrus.WithField("Region", region).Info()
		resourcesToWipe[region] = make(aws.IResourceTypeResources)

		aws.New(region, c.Config.Options.MaxRetries, c.Config.Options.RoleToAssume)
		logrus.WithField("c.Config.Filters", c.Config.Filters).Info()
		for resType, filters := range c.Config.Filters {
			rs := resourcesToWipe[region][resType]
			c.getFilteredResources(resType, filters, &rs, &warnings)
			resourcesToWipe[region][resType] = rs
		}

		logrus.WithField("Number of Resources", len(resourcesToWipe)).Info("Filtered resources")

		if c.Config.Options.DryRun == false {
			c.wipe(resourcesToWipe)
		} else {
			logrus.Info("Skip deleting resources because DryRun mode is ON")
		}
	}

	return resourcesToWipe, warnings, nil
}

func (c *Wiper) getFilteredResources(resourceType aws.ResourceType, filters filters.Filters, rs *aws.IResources, warnings *[]error) {
	logrus.WithField("Resource Type", resourceType).Info()
	logrus.WithField("Filters", filters).Info()

	if candidateResources, err := aws.List(resourceType); err != nil {
		*warnings = append(*warnings, err)
	} else {
		logrus.WithField("Number of Resources", len(candidateResources)).Debug("Got candidate resources")
		if deletableResources, err := filters.Apply(candidateResources); err != nil {
			*warnings = append(*warnings, err)
		} else {
			*rs = append(*rs, deletableResources...)
		}
	}
}

// wipe does the actual deletion (in parallel) of a given (filtered) list of AWS resources.
// (so we get retries, detaching of policies from some IAM resources before deletion, and other stuff for free).
func (c *Wiper) wipe(rtrs aws.IRegionResourceTypeResources) {
	for _, trs := range rtrs {
		for _, resources := range trs {
			for _, resource := range resources {
				if err := resource.Delete(); err != nil {
					logrus.WithError(err).WithField("Resource", resource).Error("Failed to delete a resource")
				}
			}
		}
	}
}
