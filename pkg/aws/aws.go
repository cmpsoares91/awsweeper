package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Config ...
type Config = aws.Config

// New ...
func New(config *Config) {
	sess, err := session.NewSession(config)
	if err != nil {
		fmt.Println(err)
	}

	Register(sess, &InstanceAPI{})
	Register(sess, &S3BucketAPI{})
	Register(sess, &DynamoDbTableApi{})
	Register(sess, &ElasticSearchDomainApi{})
	Register(sess, &KinesisDataStreamAPI{})
	Register(sess, &FirehoseAPI{})
}
