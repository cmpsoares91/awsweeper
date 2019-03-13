package aws

import (
	"regexp"
	"sort"
	"time"

	"github.com/iflix/awsweeper/pkg/terraform"

	"log"

	"fmt"
)

// Filter represents an entry in Config and selects the resources of a particular resource type.
type Filter struct {
	ID      *[]string `yaml:",omitempty"`
	Tags    *Tags     `yaml:",omitempty"`
	Created *Created  `yaml:",omitempty"`
	Age     *Age      `yaml:",omitempty"`
	Not     *Filter   `yaml:",omitempty`
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

// Filters selects resources based on a given yaml config.
type Filters map[terraform.ResourceType][]Filter

// Validate checks if all resource types appearing in the config are currently supported.
func (f Filters) Validate() error {
	for _, resType := range f.Types() {
		if !SupportedResourceType(resType) {
			return fmt.Errorf("unsupported resource type found in yaml config: %s", resType)
		}
	}
	return nil
}

// Types returns all the resource types in the config in their dependency order.
func (f Filters) Types() []terraform.ResourceType {
	resTypes := make([]terraform.ResourceType, 0, len(f))

	for k := range f {
		resTypes = append(resTypes, k)
	}

	sort.Slice(resTypes, func(i, j int) bool {
		return DependencyOrder[resTypes[i]] > DependencyOrder[resTypes[j]]
	})

	return resTypes
}

// MatchID checks whether a resource ID matches the filter.
func (rtf Filter) matchID(idsFilter *[]string, id string) bool {
	if idsFilter == nil {
		return true
	}

	for _, idFilter := range *idsFilter {
		if ok, err := regexp.MatchString(idFilter, id); ok {
			if err != nil {
				log.Fatal(err)
			}
			return true
		}
	}

	return false
}

// MatchesTags checks whether a resource's tags
// match the filter. The keys must match exactly, whereas the tag value is checked against a regex.
func (rtf Filter) matchTags(tagsFilter *Tags, tags map[string]string) bool {
	if tagsFilter == nil {
		return true
	}

	for _, tagFilter := range *tagsFilter {
		for cfgTagKey, regex := range tagFilter {
			if tagVal, ok := tags[cfgTagKey]; ok {
				if matched, err := regexp.MatchString(regex, tagVal); matched {
					if err != nil {
						log.Fatal(err)
					}
					return true
				}
			}
		}
	}

	return false
}

func (rtf Filter) matchCreated(createdFilter *Created, creationTime *time.Time) bool {
	if createdFilter == nil {
		return true
	}

	if creationTime == nil {
		return false
	}

	createdAfter := true
	if createdFilter.After != nil {
		createdAfter = creationTime.Unix() > createdFilter.After.Unix()
	}

	createdBefore := true
	if createdFilter.Before != nil {
		createdBefore = creationTime.Unix() < createdFilter.Before.Unix()
	}

	return createdAfter && createdBefore
}

func (rtf Filter) matchAge(ageFilter *Age, creationTime *time.Time, timeShift *time.Duration) bool {
	if ageFilter == nil {
		return true
	}

	if creationTime == nil {
		return false
	}

	now := time.Now()

	if timeShift != nil {
		now.Add(*timeShift)
	}

	createdAfter := true
	if ageFilter.YoungerThan != nil {
		createdAfter = creationTime.Unix() > now.Add(-*ageFilter.YoungerThan).Unix()
	}

	createdBefore := true
	if ageFilter.OlderThan != nil {
		createdBefore = creationTime.Unix() < now.Add(-*ageFilter.OlderThan).Unix()
	}

	return createdAfter && createdBefore
}

func (rtf Filter) match(r *Resource, timeShift *time.Duration) bool {
	matchedID := rtf.matchID(rtf.ID, r.ID)
	matchedTags := rtf.matchTags(rtf.Tags, r.Tags)
	matchedCreated := rtf.matchCreated(rtf.Created, r.Created)
	matchedAge := rtf.matchAge(rtf.Age, r.Created, timeShift)

	if rtf.Not != nil {
		if rtf.Not.ID != nil {
			matchedID = matchedID && !rtf.matchID(rtf.Not.ID, r.ID)
		}

		if rtf.Not.Tags != nil {
			matchedTags = matchedTags && !rtf.matchTags(rtf.Not.Tags, r.Tags)
		}

		if rtf.Not.Created != nil {
			matchedCreated = matchedCreated && !rtf.matchCreated(rtf.Not.Created, r.Created)
		}

		if rtf.Not.Age != nil {
			matchedAge = matchedAge && !rtf.matchAge(rtf.Not.Age, r.Created, timeShift)
		}
	}

	return matchedID &&
		matchedTags &&
		matchedCreated &&
		matchedAge
}

// matches checks whether a resource matches the filter criteria.
func (f Filters) matches(r *Resource, timeShift *time.Duration) bool {
	resFilters, found := f[r.Type]
	if !found {
		return false
	}

	if len(resFilters) == 0 {
		return true
	}

	for _, rtf := range resFilters {
		if rtf.match(r, timeShift) {
			return true
		}
	}
	return false
}
