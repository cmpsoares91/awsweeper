package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

type XYZAPI struct {
	api *s3.S3
}

func (a *XYZAPI) getType() ResourceType {
	return "aws_XYZ"
}

func (a *XYZAPI) getPriority() int64 {
	return -1
}

func (a *XYZAPI) initAPI(s *session.Session) {
	a.api = s3.New(s)
}

func (a *XYZAPI) list() (resources IResources, err error) {
	return resources, err
}

// XYZ ...
type XYZ Resource

// Delete ...
func (r *XYZ) Delete() error {
	logrus.WithField("XYZ", *r.Name).Info("Deleting XYZ")
	return nil
}

// String ...
func (r *XYZ) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *XYZ) GetID() *string { return r.ID }

// GetName ...
func (r *XYZ) GetName() *string { return r.Name }

// GetTags ...
func (r *XYZ) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *XYZ) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *XYZ) EnsureLazyLoaded() {}
