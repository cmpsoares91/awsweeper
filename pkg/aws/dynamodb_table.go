package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"
)

type DynamoDbTableApi struct {
	api *dynamodb.DynamoDB
}

func (a *DynamoDbTableApi) getType() ResourceType {
	return "dynamodb_table"
}

func (a *DynamoDbTableApi) getPriority() int64 {
	return 9710
}

func (a *DynamoDbTableApi) new(s *session.Session, cfg *aws.Config) {
	a.api = dynamodb.New(s, cfg)
}

func (a *DynamoDbTableApi) listAll(lastEvaluatedTableName *string) (tables []*string, err error) {
	if lastEvaluatedTableName != nil {
		logrus.WithField("lastEvaluatedTableName", *lastEvaluatedTableName).Warn("Listing DDB tables")
	}

	listTableOutput, err := a.api.ListTables(&dynamodb.ListTablesInput{
		ExclusiveStartTableName: lastEvaluatedTableName,
	})

	if err != nil {
		return tables, err
	}

	tables = append(tables, listTableOutput.TableNames...)

	if listTableOutput.LastEvaluatedTableName != nil {
		nextPageTables, err := a.listAll(listTableOutput.LastEvaluatedTableName)
		if err != nil {
			return tables, err
		} else {
			tables = append(tables, nextPageTables...)
		}
	}

	return tables, err
}

func (a *DynamoDbTableApi) list() (IResources, error) {
	tables, err := a.listAll(nil)
	if err != nil {
		return nil, err
	}

	var resources IResources
	for _, tableName := range tables {
		r := &DynamoDbTable{
			Name:         tableName,
			ID:           tableName,
			Tags:         make(Tags),
			ResourceType: a.getType(),
			api:          a.api,
		}

		resources = append(resources, r)
	}
	return resources, nil
}

// DynamoDbTable ...
type DynamoDbTable Resource

// Delete ...
func (r *DynamoDbTable) Delete() error {
	logrus.WithField("DDB Table", *r.Name).Info("Deleting a DDB Table")
	api := r.api.(*dynamodb.DynamoDB)
	result, err := api.DeleteTable(&dynamodb.DeleteTableInput{TableName: r.ID})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result":     result.String(),
		"Table Name": *r.Name,
	}).Info("DDB Table deleted")
	return nil
}

// String ...
func (r *DynamoDbTable) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *DynamoDbTable) GetID() *string { return r.ID }

// GetName ...
func (r *DynamoDbTable) GetName() *string { return r.Name }

// GetTags ...
func (r *DynamoDbTable) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *DynamoDbTable) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *DynamoDbTable) EnsureLazyLoaded() {
	if !r.lazyLoadPerformed {
		logrus.WithField("resource", r).Debug("Performing a lazyload on a ddb table")
		api := r.api.(*dynamodb.DynamoDB)
		if tableDesc, err := api.DescribeTable(&dynamodb.DescribeTableInput{TableName: r.ID}); err == nil {
			r.CreationDate = tableDesc.Table.CreationDateTime
			logrus.WithField("tableDesc", tableDesc).Debug("tableDesc")
			if listTagsOutput, err := api.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{ResourceArn: tableDesc.Table.TableArn}); err == nil {
				for _, tag := range listTagsOutput.Tags {
					r.Tags[*tag.Key] = *tag.Value
				}
			} else {
				logrus.WithError(err).Fatal("Failed to load ddb table Tags")
			}
		} else {
			logrus.WithError(err).Fatal("Failed to load ddb table descriptions")
		}

		r.lazyLoadPerformed = true
	}
}
