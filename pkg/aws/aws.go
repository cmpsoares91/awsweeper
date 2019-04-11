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

	register(sess, &InstanceAPI{})
	register(sess, &S3BucketAPI{})
	register(sess, &DynamoDbTableApi{})
	register(sess, &ElasticSearchDomainApi{})
	register(sess, &KinesisDataStreamAPI{})
	register(sess, &FirehoseAPI{})
}
