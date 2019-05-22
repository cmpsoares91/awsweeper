package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/sirupsen/logrus"
)

type FirehoseAPI struct {
	api *firehose.Firehose
}

func (a *FirehoseAPI) getType() ResourceType {
	return "firehose"
}

func (a *FirehoseAPI) getPriority() int64 {
	return -1
}

func (a *FirehoseAPI) new(s *session.Session, cfg *aws.Config) {
	a.api = firehose.New(s, cfg)
}

func (a *FirehoseAPI) list() (resources IResources, err error) {
	streams, err := a.api.ListDeliveryStreams(&firehose.ListDeliveryStreamsInput{})
	if err != nil {
		return nil, err
	}

	for _, deliveryStreamName := range streams.DeliveryStreamNames {
		r := &Firehose{
			Name:         deliveryStreamName,
			ID:           deliveryStreamName,
			Tags:         make(Tags),
			ResourceType: a.getType(),
			api:          a.api,
		}
		resources = append(resources, r)
	}

	return resources, err
}

// Firehose ...
type Firehose Resource

// Delete ...
func (r *Firehose) Delete() error {
	logrus.WithField("Firehose", *r.Name).Info("Deleting Firehose")
	api := r.api.(*firehose.Firehose)
	result, err := api.DeleteDeliveryStream(&firehose.DeleteDeliveryStreamInput{DeliveryStreamName: r.ID})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result": result.String(),
		"Name":   *r.Name,
	}).Info("Firehose deleted")

	return nil
}

// String ...
func (r *Firehose) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *Firehose) GetID() string {
	if r.ID != nil {
		return *r.ID
	}

	return ""
}

// GetName ...
func (r *Firehose) GetName() string {
	if r.Name != nil {
		return *r.Name
	}

	return ""
}

// GetTags ...
func (r *Firehose) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *Firehose) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *Firehose) EnsureLazyLoaded() {
	if !r.lazyLoadPerformed {
		logrus.WithField("resource", r).Debug("Performing a lazyload on a Firehose")
		api := r.api.(*firehose.Firehose)

		tagsOutput, err := api.ListTagsForDeliveryStream(&firehose.ListTagsForDeliveryStreamInput{DeliveryStreamName: r.ID})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load Firehose tags")
		}

		if tagsOutput.Tags != nil {
			for _, tag := range tagsOutput.Tags {
				r.Tags[*tag.Key] = *tag.Value
			}
		}

		descStream, err := api.DescribeDeliveryStream(&firehose.DescribeDeliveryStreamInput{DeliveryStreamName: r.ID})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load Firehose description")
		}

		r.CreationDate = descStream.DeliveryStreamDescription.CreateTimestamp
		r.lazyLoadPerformed = true
	}
}
