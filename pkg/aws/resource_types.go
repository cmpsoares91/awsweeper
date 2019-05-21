package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

type ResourceType string

type iResourceType interface {
	new(*session.Session, *aws.Config)
	list() (IResources, error)
	getType() ResourceType
	getPriority() int64
}

var registeredResourceTypes = make(map[ResourceType]iResourceType)

// IsRegistered ...
func IsRegistered(resourceType ResourceType) bool {
	logrus.WithField("resourceType", resourceType).Debug("Checking if resourceType is supported")

	if _, ok := registeredResourceTypes[resourceType]; ok {
		return true
	}

	return false
}

// Register ...
func register(s *session.Session, cfg *aws.Config, r iResourceType) {
	logrus.WithField("ResourceType", r.getType()).Debug("Registering new resource type")
	r.new(s, cfg)
	registeredResourceTypes[r.getType()] = r
}

// List ...
func List(resourceType ResourceType) (IResources, error) {
	if !IsRegistered(resourceType) {
		return nil, fmt.Errorf("ResourceType (%v) is not supported", resourceType)
	}

	return registeredResourceTypes[resourceType].list()
}
