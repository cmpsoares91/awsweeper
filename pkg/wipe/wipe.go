package wipe

import (
	"fmt"
	"sync"

	"log"

	"github.com/hashicorp/terraform/terraform"
	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/iflix/awsweeper/pkg/config"
	"github.com/sirupsen/logrus"
)

// Wiper is currently the only command.
//
// It deletes selected AWS resources by a given filter
type Wiper struct {
	DryRun   bool
	Client   *aws.API
	Provider *terraform.ResourceProvider
	Filters  *config.Config
}

func (c *Wiper) Run() (aws.Resources, error) {
	err := c.Filters.Validate()
	if err != nil {
		return nil, err
	}

	if c.Client == nil {
		return nil, fmt.Errorf("AWS Client is required")
	}

	if c.Provider == nil {
		return nil, fmt.Errorf("Provider is required")
	}

	if c.DryRun {
		logrus.Info("This is a test run, nothing will be deleted!")
	}

	var resourcesToWipe = aws.Resources{}
	for _, resType := range c.Filters.Types() {
		rawResources, err := c.Client.RawResources(resType)
		if err != nil {
			return nil, err
		}

		deletableResources, err := aws.DeletableResources(resType, rawResources)
		if err != nil {
			return nil, err
		}

		filteredRes := c.Filters.Apply(resType, deletableResources, rawResources, c.Client)
		resourcesToWipe = append(resourcesToWipe, filteredRes...)

		if c.DryRun == false {
			c.wipe(filteredRes)
		}
	}

	return resourcesToWipe, nil
}

// wipe does the actual deletion (in parallel) of a given (filtered) list of AWS resources.
// It takes advantage of the AWS terraform provider by using its delete functions
// (so we get retries, detaching of policies from some IAM resources before deletion, and other stuff for free).
func (c *Wiper) wipe(res aws.Resources) {
	numWorkerThreads := 10

	if len(res) == 0 {
		return
	}

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

					state, err := (*c.Provider).Refresh(instanceInfo, s)
					if err != nil {
						log.Fatal(err)
					}

					// doesn't hurt to always add some force attributes
					state.Attributes["force_detach_policies"] = "true"
					state.Attributes["force_destroy"] = "true"

					_, err = (*c.Provider).Apply(instanceInfo, state, instanceDiff)

					if err != nil {
						logrus.Error(err)
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
