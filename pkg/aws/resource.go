package aws

import (
	"time"
	"fmt"
)

// Region ...
type Region = string

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

// IResource ...
type IResource interface {
	GetID() string
	GetName() string
	GetTags() *Tags
	GetCreationDate() *time.Time
	Delete() error
	String() string
	EnsureLazyLoaded()
}

// IResources ...
type IResources []IResource

// IResourceTypeResources ...
type IResourceTypeResources map[ResourceType]IResources

// IRegionResourceTypeResources ...
type IRegionResourceTypeResources map[Region]IResourceTypeResources

func (rrtrs *IRegionResourceTypeResources) String() string {
	output := ""
	for region, rtrs := range *rrtrs {
		for resourceType, resources := range rtrs {
			for _, r := range resources {
				output = output + fmt.Sprintf("- [%s][%s][%s] %s\n", region, resourceType, r.GetID(), r.GetName())
			}
		}
	}

	return output
}

func (rrtrs *IRegionResourceTypeResources) Len() int {
	counter := 0
	for _, rtrs := range *rrtrs {
		for _, resources := range rtrs {
			counter += len(resources)
		}
	}

	return counter
}