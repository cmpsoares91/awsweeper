package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

// New ...
func New(region string, maxRetries int, roleToAssume string) {
	config := &aws.Config{
		Region:     &region,
		MaxRetries: &maxRetries,
	}

	sess, err := session.NewSession(config)
	if err != nil {
		fmt.Println(err)
	}

	if roleToAssume != "" {
		logrus.WithField("Role", roleToAssume).Info("Assuming Role")
		config.Credentials = stscreds.NewCredentials(sess, roleToAssume)
	}

	register(sess, config, &EC2API{})
	register(sess, config, &S3BucketAPI{})
	register(sess, config, &DynamoDbTableApi{})
	register(sess, config, &ElasticSearchDomainApi{})
	register(sess, config, &KinesisDataStreamAPI{})
	register(sess, config, &FirehoseAPI{})
}
