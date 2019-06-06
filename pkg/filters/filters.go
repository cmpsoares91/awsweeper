package filters

import (
	"fmt"
	"strings"
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
		"Filter": filter,
		"Number of Resources Before Filtering": len(resources),
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

func (filter Filter) String() string {
	var output []string
	if filter.IDs != nil {
		output = append(output, fmt.Sprintf("IDS:[%s]", strings.Join(*filter.IDs, ",")))
	}
	if filter.Tags != nil {
		var ts []string
		for _, t := range *filter.Tags {
			tss := "("
			for k, v := range t {
				tss = tss + fmt.Sprintf("%s=%v,", k, v)
			}
			ts = append(ts, tss+")")
		}
		output = append(output, fmt.Sprintf("TAGS:[%s]", strings.Join(ts, ",")))
	}

	if filter.Created != nil {
		if filter.Created.Before != nil {
			output = append(output, fmt.Sprintf("BEFORE:[%s]", filter.Created.Before.String()))
		}
		if filter.Created.After != nil {
			output = append(output, fmt.Sprintf("AFTER:[%s]", filter.Created.After.String()))
		}
	}

	if filter.Age != nil {
		if filter.Age.YoungerThan != nil {
			output = append(output, fmt.Sprintf("YOUNGERTHAN:[%s]", filter.Age.YoungerThan.String()))
		}
		if filter.Age.OlderThan != nil {
			output = append(output, fmt.Sprintf("OLDERTHAN:[%s]", filter.Age.OlderThan.String()))
		}
	}

	if filter.Not != nil {
		var ns []string
		for _, fn := range *filter.Not {
			ns = append(ns, fn.String())
		}
		output = append(output, fmt.Sprintf("NOT:{%s}", strings.Join(ns, ",")))
	}

	return strings.Join(output, ", ")
}
