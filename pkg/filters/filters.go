package filters

import (
	"time"

	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/sirupsen/logrus"
)

type Filters []Filter

// Filter represents an entry in Config and selects the resources of a particular resource type.
type Filter struct {
	IDs     *[]string `yaml:",omitempty"`
	Tags    *Tags     `yaml:",omitempty"`
	Created *Created  `yaml:",omitempty"`
	Age     *Age      `yaml:",omitempty"`
	Not     *Filters  `yaml:",omitempty`
}

type Tags []map[string]string

type Created struct {
	Before *time.Time `yaml:",omitempty"`
	After  *time.Time `yaml:",omitempty"`
}

type Age struct {
	OlderThan   *time.Duration `yaml:"older_than,omitempty"`
	YoungerThan *time.Duration `yaml:"younger_than,omitempty"`
}

func (filters Filters) Apply(resources aws.IResources) (filteredResources aws.IResources, err error) {
	logrus.WithField("Number of filters", len(filters)).Debug("Applying Filters")

	if len(filters) == 0 {
		return resources, err
	}

	for _, filter := range filters {
		fr, err := filter.Apply(resources)
		if err != nil {
			return nil, err
		}

		filteredResources = append(filteredResources, fr...)
	}

	return filteredResources, err
}

func (filter Filter) Apply(resources aws.IResources) (filteredResources aws.IResources, err error) {
	logrus.WithFields(logrus.Fields{
		"Filter":              filter,
		"Number of Resources": len(resources),
	}).Info("Apply Filter")

	filteredResources = resources
	filteredResources, err = filter.byIDs(filteredResources)
	if err != nil {
		return nil, err
	}

	filteredResources, err = filter.byTags(filteredResources)
	if err != nil {
		return nil, err
	}

	filteredResources, err = filter.byCreated(filteredResources)
	if err != nil {
		return nil, err
	}

	filteredResources, err = filter.byAge(filteredResources)
	if err != nil {
		return nil, err
	}

	filteredResources, err = filter.byNot(filteredResources)
	if err != nil {
		return nil, err
	}

	return filteredResources, err
}
