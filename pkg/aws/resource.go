package aws

import (
	"time"
)

// Resource ...
type Resource struct {
	Name              *string
	ID                *string
	Tags              Tags
	CreationDate      *time.Time
	ResourceType      ResourceType
	api               interface{}
	lazyLoadPerformed bool
}

// Tags ...
type Tags map[string]string

// IResource
type IResource interface {
	GetID() *string
	GetName() *string
	GetTags() *Tags
	GetCreationDate() *time.Time
	Delete() error
	String() string
	EnsureLazyLoaded()
}

// IResources ...
type IResources []IResource
