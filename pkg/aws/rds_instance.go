package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

type RDSInstanceAPI struct {
	api *rds.RDS
}

func (a *RDSInstanceAPI) getType() ResourceType {
	return "rds_instance"
}

func (a *RDSInstanceAPI) getPriority() int64 {
	return -1
}

func (a *RDSInstanceAPI) new(s *session.Session, cfg *aws.Config) {
	a.api = rds.New(s, cfg)
}

func (a *RDSInstanceAPI) list() (resources IResources, err error) {
	output, err := a.api.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, err
	}

	for _, instance := range output.DBInstances {
		if instance.DBClusterIdentifier == nil {
			r := &RDSInstance{
				Name:         instance.DBInstanceArn,
				ID:           instance.DBInstanceIdentifier,
				CreationDate: instance.InstanceCreateTime,
				Tags:         make(Tags),
				ResourceType: a.getType(),
				api:          a.api,
			}
			resources = append(resources, r)
		} else {
			logrus.WithField("DBClusterIdentifier", *instance.DBClusterIdentifier).Info("Ignoring RdsInstance because it is part of cluster")
		}
	}

	return resources, err
}

// XYZ ...
type RDSInstance Resource

// Delete ...
func (r *RDSInstance) Delete() error {
	logrus.WithField("RDSInstance", *r.Name).Info("Deleting RDSInstance")
	api := r.api.(*rds.RDS)

	_, err := api.ModifyDBInstance(&rds.ModifyDBInstanceInput{
		ApplyImmediately:     aws.Bool(true),
		DeletionProtection:   aws.Bool(false),
		DBInstanceIdentifier: r.ID,
	})
	if err != nil {
		return err
	}

	result, err := api.DeleteDBInstance(&rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: r.ID,
		SkipFinalSnapshot:    aws.Bool(true),
	})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result": result.String(),
		"Name":   *r.Name,
	}).Info("RDSInstance deleted")

	return nil
}

// String ...
func (r *RDSInstance) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *RDSInstance) GetID() string {
	if r.ID != nil {
		return *r.ID
	}

	return ""
}

// GetName ...
func (r *RDSInstance) GetName() string {
	if r.Name != nil {
		return *r.Name
	}

	return ""
}

// GetTags ...
func (r *RDSInstance) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *RDSInstance) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *RDSInstance) EnsureLazyLoaded() {
	if !r.lazyLoadPerformed {
		logrus.WithField("resource", r).Debug("Performing a lazyload on a RDSInstance")
		api := r.api.(*rds.RDS)

		tagsOutput, err := api.ListTagsForResource(&rds.ListTagsForResourceInput{ResourceName: r.Name})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load RdsInstance tags")
		}

		if tagsOutput.TagList != nil {
			for _, tag := range tagsOutput.TagList {
				r.Tags[*tag.Key] = *tag.Value
			}
		}

		r.lazyLoadPerformed = true
	}
}
