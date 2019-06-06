package aws

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

type RDSClusterAPI struct {
	api *rds.RDS
}

func (a *RDSClusterAPI) getType() ResourceType {
	return "rds_cluster"
}

func (a *RDSClusterAPI) getPriority() int64 {
	return -1
}

func (a *RDSClusterAPI) new(s *session.Session, cfg *aws.Config) {
	a.api = rds.New(s, cfg)
}

func (a *RDSClusterAPI) list() (resources IResources, err error) {
	output, err := a.api.DescribeDBClusters(&rds.DescribeDBClustersInput{})
	if err != nil {
		return nil, err
	}

	for _, cluster := range output.DBClusters {
		r := &RDSCluster{
			Name:         cluster.DBClusterArn,
			ID:           cluster.DBClusterIdentifier,
			CreationDate: cluster.ClusterCreateTime,
			Tags:         make(Tags),
			ResourceType: a.getType(),
			api:          a.api,
		}
		resources = append(resources, r)
	}

	return resources, err
}

// RDSCluster ...
type RDSCluster Resource

// Delete ...
func (r *RDSCluster) Delete() error {
	logrus.WithField("RDSCluster", *r.Name).Info("Deleting RDSCluster")
	api := r.api.(*rds.RDS)

	_, err := api.ModifyDBCluster(&rds.ModifyDBClusterInput{
		ApplyImmediately:    aws.Bool(true),
		DeletionProtection:  aws.Bool(false),
		DBClusterIdentifier: r.ID,
	})
	if err != nil {
		return err
	}

	output, err := api.DescribeDBClusters(&rds.DescribeDBClustersInput{DBClusterIdentifier: r.ID})
	if err != nil {
		return err
	}

	for _, instance := range output.DBClusters[0].DBClusterMembers {
		logrus.WithField("ID", *instance.DBInstanceIdentifier).Info("Deleting cluster member")
		_, err := api.DeleteDBInstance(&rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: instance.DBInstanceIdentifier,
			SkipFinalSnapshot:    aws.Bool(true),
		})
		if err != nil {
			return err
		}
	}

	result, err := api.DeleteDBCluster(&rds.DeleteDBClusterInput{
		DBClusterIdentifier:       r.ID,
		SkipFinalSnapshot:         aws.Bool(false),
		FinalDBSnapshotIdentifier: aws.String(fmt.Sprintf("%s-final-snapshot", *r.ID)),
	})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result": result.String(),
		"Name":   *r.Name,
	}).Info("RDSCluster deleted")

	return nil
}

// String ...
func (r *RDSCluster) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *RDSCluster) GetID() string {
	if r.ID != nil {
		return *r.ID
	}

	return ""
}

// GetName ...
func (r *RDSCluster) GetName() string {
	if r.Name != nil {
		return *r.Name
	}

	return ""
}

// GetTags ...
func (r *RDSCluster) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *RDSCluster) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *RDSCluster) EnsureLazyLoaded() {
	if !r.lazyLoadPerformed {
		logrus.WithField("resource", r).Debug("Performing a lazyload on a RDSCluster")
		api := r.api.(*rds.RDS)

		tagsOutput, err := api.ListTagsForResource(&rds.ListTagsForResourceInput{ResourceName: r.Name})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load RDSCluster tags")
		}

		if tagsOutput.TagList != nil {
			for _, tag := range tagsOutput.TagList {
				r.Tags[*tag.Key] = *tag.Value
			}
		}

		r.lazyLoadPerformed = true
	}
}
