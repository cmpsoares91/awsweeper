package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/sirupsen/logrus"
)

type MediaLiveInputAPI struct {
	api *medialive.MediaLive
}

func (a *MediaLiveInputAPI) getType() ResourceType {
	return "medialive_input"
}

func (a *MediaLiveInputAPI) getPriority() int64 {
	return -1
}

func (a *MediaLiveInputAPI) new(s *session.Session, cfg *aws.Config) {
	a.api = medialive.New(s, cfg)
}

func (a *MediaLiveInputAPI) list() (resources IResources, err error) {
	list, err := a.api.ListInputs(&medialive.ListInputsInput{})
	if err != nil {
		return nil, err
	}

	for _, input := range list.Inputs {
		r := &MediaLiveInput{
			Name:         input.Name,
			ID:           input.Id,
			Tags:         make(Tags),
			CreationDate: nil,
			ResourceType: a.getType(),
			api:          a.api,
		}

		for k, v := range input.Tags {
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
			input.Tags[FirstSeenDateTimeMarker] = aws.String(time.Now().Format(time.RFC3339))
			_, err := a.api.CreateTags(&medialive.CreateTagsInput{
				ResourceArn: input.Arn,
				Tags:        input.Tags,
			})

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"Error": err.Error(),
					"ID":    *input.Id,
					"Name":  *input.Name,
				}).Warn("Failed to set marker tag for a resource")
			}
		}

		resources = append(resources, r)
	}

	return resources, err
}

// MediaLiveInput ...
type MediaLiveInput Resource

// Delete ...
func (r *MediaLiveInput) Delete() error {
	logrus.WithField("MediaLiveInput", *r.Name).Info("Deleting MediaLiveInput")
	api := r.api.(*medialive.MediaLive)
	result, err := api.DeleteInput(&medialive.DeleteInputInput{InputId: r.ID})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result": result.String(),
		"Name":   *r.Name,
	}).Info("MediaLiveInput deleted")

	return nil
}

// String ...
func (r *MediaLiveInput) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *MediaLiveInput) GetID() string {
	if r.ID != nil {
		return *r.ID
	}

	return ""
}

// GetName ...
func (r *MediaLiveInput) GetName() string {
	if r.Name != nil {
		return *r.Name
	}

	return ""
}

// GetTags ...
func (r *MediaLiveInput) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *MediaLiveInput) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *MediaLiveInput) EnsureLazyLoaded() {}
