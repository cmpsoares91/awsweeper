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

func (c *Wiper) Run() (aws.IResources, []error, error) {
	var warnings []error
	var resourcesToWipe aws.IResources

	for _, region := range c.Config.Options.Regions {
		logrus.WithField("Region", region).Info()
		aws.New(&aws.Config{
			Region:           &region,
			MaxRetries:       &c.Config.Options.MaxRetries,
			S3ForcePathStyle: &c.Config.Options.S3ForcePathStyle,
		})

		logrus.WithField("c.Config.Filters", c.Config.Filters).Info()
		for resType, filters := range c.Config.Filters {
			c.getFilteredResources(resType, filters, &resourcesToWipe, &warnings)
		}

		logrus.WithField("Number of Resources", len(resourcesToWipe)).Info("Filtered resources")

		if c.Config.Options.DryRun == false {
			logrus.Info("DryRun mode is OFF")
			c.wipe(resourcesToWipe)
		} else {
			logrus.Info("DryRun mode is ON. Skip deleting resources")
		}
	}

	return resourcesToWipe, warnings, nil
}

func (c *Wiper) getFilteredResources(resourceType aws.ResourceType, filters filters.Filters, resourcesToWipe *aws.IResources, warnings *[]error) {
	logrus.WithField("Resource Type", resourceType).Info()
	logrus.WithField("Filters", filters).Info()

	if candidateResources, err := aws.List(resourceType); err != nil {
		*warnings = append(*warnings, err)
	} else {
		logrus.WithField("Number of Resources", len(candidateResources)).Debug("Got candidate resources")
		if deletableResources, err := filters.Apply(candidateResources); err != nil {
			*warnings = append(*warnings, err)
		} else {
			*resourcesToWipe = append(*resourcesToWipe, deletableResources...)
		}
	}
}

// wipe does the actual deletion (in parallel) of a given (filtered) list of AWS resources.
// (so we get retries, detaching of policies from some IAM resources before deletion, and other stuff for free).
func (c *Wiper) wipe(resources aws.IResources) {
	for _, resource := range resources {
		if err := resource.Delete(); err != nil {
			logrus.WithError(err).WithField("Resource", resource).Error("Failed to delete a resource")
		}
	}
}
