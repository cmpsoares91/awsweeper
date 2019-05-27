package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/sirupsen/logrus"
)

type MediaLiveChannelAPI struct {
	api *medialive.MediaLive
}

func (a *MediaLiveChannelAPI) getType() ResourceType {
	return "medialive_channel"
}

func (a *MediaLiveChannelAPI) getPriority() int64 {
	return -1
}

func (a *MediaLiveChannelAPI) new(s *session.Session, cfg *aws.Config) {
	a.api = medialive.New(s, cfg)
}

func (a *MediaLiveChannelAPI) list() (resources IResources, err error) {
	list, err := a.api.ListChannels(&medialive.ListChannelsInput{})
	if err != nil {
		return nil, err
	}

	for _, channel := range list.Channels {
		r := &MediaLiveChannel{
			Name:         channel.Name,
			ID:           channel.Id,
			Tags:         make(Tags),
			CreationDate: nil,
			ResourceType: a.getType(),
			api:          a.api,
		}

		for k, v := range channel.Tags {
			if r.CreationDate == nil && k == FirstSeenDateTimeMarker {
				firstSeenDate, err := time.Parse(time.RFC3339, *v)
				if err != nil {
					logrus.WithField("TagValue", *v).Warn("Failed to parse marker tag value into DateTime")
				} else {
					r.CreationDate = &firstSeenDate
				}
			} else {
				r.Tags[k] = *v
			}
		}

		// Since medialive resources are not providing creationDate, we'll utilize tags
		// to mark the resources first time we see them and then use tag marker as createdDate
		if r.CreationDate == nil {
			channel.Tags[FirstSeenDateTimeMarker] = aws.String(time.Now().Format(time.RFC3339))
			_, err := a.api.CreateTags(&medialive.CreateTagsInput{
				ResourceArn: channel.Arn,
				Tags:        channel.Tags,
			})

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Error": err.Error(),
					"ID":    *channel.Id,
					"Name":  *channel.Name,
				}).Warn("Failed to set marker tag for a medialive_channel resource")
			}
		}

		resources = append(resources, r)
	}

	return resources, err
}

// MediaLiveChannel ...
type MediaLiveChannel Resource

// Delete ...
func (r *MediaLiveChannel) Delete() error {
	logrus.WithField("MediaLiveChannel", *r.Name).Info("Deleting MediaLiveChannel")
	api := r.api.(*medialive.MediaLive)
	result, err := api.DeleteInput(&medialive.DeleteInputInput{InputId: r.ID})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result": result.String(),
		"Name":   *r.Name,
	}).Info("MediaLiveChannel deleted")

	return nil
}

// String ...
func (r *MediaLiveChannel) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *MediaLiveChannel) GetID() string {
	if r.ID != nil {
		return *r.ID
	}

	return ""
}

// GetName ...
func (r *MediaLiveChannel) GetName() string {
	if r.Name != nil {
		return *r.Name
	}

	return ""
}

// GetTags ...
func (r *MediaLiveChannel) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *MediaLiveChannel) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *MediaLiveChannel) EnsureLazyLoaded() {}
