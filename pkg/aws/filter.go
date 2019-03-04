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
	ID      *string           `yaml:",omitempty"`
	Tags    map[string]string `yaml:",omitempty"`
	Created *Created          `yaml:",omitempty"`
	Age     *Age              `yaml:",omitempty"`
}

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
func (rtf Filter) matchID(id string) bool {
	if rtf.ID == nil {
		return true
	}

	if ok, err := regexp.MatchString(*rtf.ID, id); ok {
		if err != nil {
			log.Fatal(err)
		}
		return true
	}

	return false
}

// MatchesTags checks whether a resource's tags
// match the filter. The keys must match exactly, whereas the tag value is checked against a regex.
func (rtf Filter) matchTags(tags map[string]string) bool {
	if rtf.Tags == nil {
		return true
	}

	for cfgTagKey, regex := range rtf.Tags {
		if tagVal, ok := tags[cfgTagKey]; ok {
			if matched, err := regexp.MatchString(regex, tagVal); !matched {
				if err != nil {
					log.Fatal(err)
				}
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (rtf Filter) matchCreated(creationTime *time.Time) bool {
	if rtf.Created == nil {
		return true
	}

	if creationTime == nil {
		return false
	}

	createdAfter := true
	if rtf.Created.After != nil {
		createdAfter = creationTime.Unix() > rtf.Created.After.Unix()
	}

	createdBefore := true
	if rtf.Created.Before != nil {
		createdBefore = creationTime.Unix() < rtf.Created.Before.Unix()
	}

	return createdAfter && createdBefore
}

func (rtf Filter) matchAge(creationTime *time.Time) bool {
	if rtf.Age == nil {
		return true
	}

	if creationTime == nil {
		return false
	}

	now := time.Now()
	createdAfter := true
	if rtf.Age.YoungerThan != nil {
		createdAfter = creationTime.Unix() > now.Add(-*rtf.Age.YoungerThan).Unix()
	}

	createdBefore := true
	if rtf.Age.OlderThan != nil {
		createdBefore = creationTime.Unix() < now.Add(-*rtf.Age.OlderThan).Unix()
	}

	return createdAfter && createdBefore
}

// matches checks whether a resource matches the filter criteria.
func (f Filters) matches(r *Resource) bool {
	resFilters, found := f[r.Type]
	if !found {
		return false
	}

	if len(resFilters) == 0 {
		return true
	}

	for _, rtf := range resFilters {
		if rtf.matchTags(r.Tags) && rtf.matchID(r.ID) && rtf.matchCreated(r.Created) && rtf.matchAge(r.Created) {
			return true
		}
	}
	return false
}
