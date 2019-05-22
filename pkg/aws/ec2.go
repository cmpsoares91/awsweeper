package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

// EC2API ...
type EC2API struct {
	api *ec2.EC2
}

func (a *EC2API) getType() ResourceType {
	return "ec2"
}

func (a *EC2API) getPriority() int64 {
	return 9980
}

func (a *EC2API) new(s *session.Session, cfg *aws.Config) {
	a.api = ec2.New(s, cfg)
}

func (a *EC2API) list() (IResources, error) {
	output, err := a.api.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("pending"), aws.String("running"),
					aws.String("stopping"), aws.String("stopped"),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	var resources IResources
	for _, rsv := range output.Reservations {
		for _, instance := range rsv.Instances {
			var resource = &Instance{
				Name:         nil,
				ID:           instance.InstanceId,
				Tags:         make(Tags),
				CreationDate: instance.LaunchTime,
				ResourceType: a.getType(),
				api:          a.api,
			}

			for _, tag := range instance.Tags {
				resource.Tags[*tag.Key] = *tag.Value

				if *tag.Key == "Name" {
					resource.Name = tag.Value
				}
			}

			resources = append(resources, resource)
		}

	}

	return resources, nil
}

// Instance ...
type Instance Resource

// GetID ...
func (r *Instance) GetID() string {
	if r.ID != nil {
		return *r.ID
	}

	return ""
}

// GetName ...
func (r *Instance) GetName() string {
	if r.Name != nil {
		return *r.Name
	}

	return ""
}

// GetTags ...
func (r *Instance) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *Instance) GetCreationDate() *time.Time { return r.CreationDate }

// Delete ...
func (r *Instance) Delete() error {
	logrus.WithField("EC2", *r.Name).Info("Deleting an EC2")
	api := r.api.(*ec2.EC2)

	result, err := api.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{r.ID},
	})

	if err != nil {
		return err
	}

	if result.TerminatingInstances != nil {
		for _, ti := range result.TerminatingInstances {
			logrus.WithFields(logrus.Fields{
				"CurrentState":  *ti.CurrentState,
				"PreviousState": *ti.PreviousState,
				"InstanceId":    *ti.InstanceId,
			}).Info("Instance is terminating")
		}
	}

	return nil
}

// String ...
func (r *Instance) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// EnsureLazyLoaded ...
func (r *Instance) EnsureLazyLoaded() {}
