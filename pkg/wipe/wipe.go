package wipe

import (
	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/iflix/awsweeper/pkg/config"
	"github.com/mousavian/limiter"
	"github.com/sirupsen/logrus"
)

type Wiper struct {
	Config *config.Config
}

func (c *Wiper) Run() (aws.IResources, []error, error) {
	var warnings []error
	aws.New(&aws.Config{
		Region:           &c.Config.Options.Regions[0],
		MaxRetries:       &c.Config.Options.MaxRetries,
		S3ForcePathStyle: &c.Config.Options.S3ForcePathStyle,
	})

	logrus.WithField("c.Config.Filters", c.Config.Filters).Info()
	var resourcesToWipe aws.IResources
	for resType, filters := range c.Config.Filters {
		logrus.WithField("resType", resType).Info()
		logrus.WithField("filters", filters).Info()
		if candidateResources, err := aws.List(resType); err != nil {
			warnings = append(warnings, err)
		} else {
			logrus.WithField("Number of Candidate Resources", len(candidateResources)).Debug("got candidateResources")
			if deletableResources, err := filters.Apply(candidateResources); err != nil {
				warnings = append(warnings, err)
			} else {
				resourcesToWipe = append(resourcesToWipe, deletableResources...)
			}
		}
	}

	logrus.WithField("numberOfResources", len(resourcesToWipe)).Info("Filtered some resources")

	if c.Config.Options.DryRun == false {
		logrus.Info("DryRun mode is OFF")
		c.wipe(resourcesToWipe)
	} else {
		logrus.Info("DryRun mode is ON. Skip deleting resources")
	}

	return resourcesToWipe, warnings, nil
}

// wipe does the actual deletion (in parallel) of a given (filtered) list of AWS resources.
// (so we get retries, detaching of policies from some IAM resources before deletion, and other stuff for free).
func (c *Wiper) wipe(resources aws.IResources) {
	var goroutineLimit = limiter.NewConcurrencyLimiter(5)

	for _, r := range resources {
		goroutineLimit.ExecuteWithParams(func(params ...interface{}) {
			resource := params[0].(aws.IResource)
			err := resource.Delete()
			if err != nil {
				logrus.WithError(err).WithField("Resource", resource).Fatal("Failed to delete a resource")
			}
		}, r)
	}

	goroutineLimit.Wait()
}
