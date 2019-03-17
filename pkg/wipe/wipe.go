package wipe

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/terraform/terraform"
	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/iflix/awsweeper/pkg/config"
	"github.com/sirupsen/logrus"
)

// Wiper is currently the only command.
//
// It deletes selected AWS resources by a given filter
type Wiper struct {
	DryRun    bool
	Client    *aws.API
	Provider  *terraform.ResourceProvider
	Filters   *config.Config
	TimeShift *time.Duration
}

func (c *Wiper) Run() (aws.Resources, []error, error) {
	var warnings []error
	err := c.Filters.Validate()
	if err != nil {
		return nil, nil, err
	}

	if c.Client == nil {
		return nil, warnings, fmt.Errorf("AWS Client is required")
	}

	if c.Provider == nil {
		return nil, warnings, fmt.Errorf("Provider is required")
	}

	if c.DryRun {
		logrus.Info("This is a test run, nothing will be deleted!")
	}

	var resourcesToWipe = aws.Resources{}
	for _, resType := range c.Filters.Types() {
		if rawResources, err := c.Client.RawResources(resType); err != nil {
			warnings = append(warnings, err)
		} else {
			if deletableResources, err := aws.DeletableResources(resType, rawResources); err != nil {
				warnings = append(warnings, err)
			} else {
				filteredRes := c.Filters.Apply(resType, deletableResources, rawResources, c.Client, c.TimeShift)
				resourcesToWipe = append(resourcesToWipe, filteredRes...)

			}
		}
	}

	logrus.WithField("numberOfResources", len(resourcesToWipe)).Info("Filtered some resources")

	if c.DryRun == false {
		c.wipe(resourcesToWipe)
	}

	return resourcesToWipe, warnings, nil
}

// wipe does the actual deletion (in parallel) of a given (filtered) list of AWS resources.
// It takes advantage of the AWS terraform provider by using its delete functions
// (so we get retries, detaching of policies from some IAM resources before deletion, and other stuff for free).
func (c *Wiper) wipe(res aws.Resources) {
	numWorkerThreads := 1

	if len(res) == 0 {
		return
	}

	logrus.WithFields(logrus.Fields{
		"numWorkerThreads": numWorkerThreads,
		"resources":        res,
	}).Info("Going to wipe some resources")

	instanceDiff := &terraform.InstanceDiff{
		Destroy: true,
	}

	chResources := make(chan *aws.Resource, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(res))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				r, more := <-chResources

				if more {
					// dirty hack to fix aws_key_pair
					if r.Attrs == nil {
						r.Attrs = map[string]string{"public_key": ""}
					}

					instanceInfo := &terraform.InstanceInfo{
						Type: string(r.Type),
					}
					s := &terraform.InstanceState{
						ID:         r.ID,
						Attributes: r.Attrs,
					}

					logrus.WithFields(logrus.Fields{
						"instanceInfo": instanceInfo,
						"state":        s,
					}).Info("Refreshing state")

					state, err := (*c.Provider).Refresh(instanceInfo, s)
					if err != nil {
						logrus.WithError(err).Fatal("Unable to refresh instance info")
					}

					// doesn't hurt to always add some force attributes
					state.Attributes["force_detach_policies"] = "true"
					state.Attributes["force_destroy"] = "true"

					logrus.WithFields(logrus.Fields{
						"instanceInfo": instanceInfo,
						"state":        state,
						"instanceDiff": instanceDiff,
					}).Info("Applying new state")

					_, err = (*c.Provider).Apply(instanceInfo, state, instanceDiff)

					if err != nil {
						logrus.WithError(err).Error("Unable to apply new state")
					}
					wg.Done()
				} else {
					return
				}
			}
		}()
	}

	for _, r := range res {
		chResources <- r
	}
	close(chResources)

	wg.Wait()
}
