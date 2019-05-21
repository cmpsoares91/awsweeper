package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/sirupsen/logrus"
)

type KinesisDataStreamAPI struct {
	api *kinesis.Kinesis
}

func (a *KinesisDataStreamAPI) getType() ResourceType {
	return "aws_kinesis_data_stream"
}

func (a *KinesisDataStreamAPI) getPriority() int64 {
	return -1
}

func (a *KinesisDataStreamAPI) new(s *session.Session, cfg *aws.Config) {
	a.api = kinesis.New(s, cfg)
}

func (a *KinesisDataStreamAPI) list() (resources IResources, err error) {
	streams, err := a.api.ListStreams(&kinesis.ListStreamsInput{})
	if err != nil {
		return nil, err
	}

	for _, streamName := range streams.StreamNames {
		r := &KinesisDataStream{
			Name:         streamName,
			ID:           streamName,
			Tags:         make(Tags),
			ResourceType: a.getType(),
			api:          a.api,
		}
		resources = append(resources, r)
	}

	return resources, err
}

// KinesisDataStream ...
type KinesisDataStream Resource

// Delete ...
func (r *KinesisDataStream) Delete() error {
	logrus.WithField("KinesisDataStream", *r.Name).Info("Deleting KinesisDataStream")
	api := r.api.(*kinesis.Kinesis)
	result, err := api.DeleteStream(&kinesis.DeleteStreamInput{StreamName: r.ID})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result":      result.String(),
		"Dtream Name": *r.Name,
	}).Info("KinesisDataStream deleted")

	return nil
}

// String ...
func (r *KinesisDataStream) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *KinesisDataStream) GetID() *string { return r.ID }

// GetName ...
func (r *KinesisDataStream) GetName() *string { return r.Name }

// GetTags ...
func (r *KinesisDataStream) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *KinesisDataStream) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *KinesisDataStream) EnsureLazyLoaded() {
	if !r.lazyLoadPerformed {
		logrus.WithField("resource", r).Debug("Performing a lazyload on a KinesisDataStream")
		api := r.api.(*kinesis.Kinesis)

		tagsOutput, err := api.ListTagsForStream(&kinesis.ListTagsForStreamInput{StreamName: r.ID})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load KinesisDataStream tags")
		}

		if tagsOutput.Tags != nil {
			for _, tag := range tagsOutput.Tags {
				r.Tags[*tag.Key] = *tag.Value
			}
		}

		descStream, err := api.DescribeStream(&kinesis.DescribeStreamInput{StreamName: r.ID})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load KinesisDataStream description")
		}

		r.CreationDate = descStream.StreamDescription.StreamCreationTimestamp
		r.lazyLoadPerformed = true
	}
}
