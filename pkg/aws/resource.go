package aws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/iflix/awsweeper/pkg/terraform"
	"github.com/pkg/errors"
)

// Resources is a list of AWS resources.
type Resources []*Resource

// Resource contains information about a single AWS resource that can be deleted by Terraform.
type Resource struct {
	Type terraform.ResourceType
	// ID by which the resource can be deleted (in some cases the ID is the resource's name, but not always;
	// that's why we need the deleteIDs map)
	ID      string
	Tags    map[string]string
	Created *time.Time
	Attrs   map[string]string
}

// Resources converts given raw resources for a given resource type
// into a format that can be deleted by the Terraform API.
func DeletableResources(resType terraform.ResourceType, resources interface{}) (Resources, error) {
	deletableResources := Resources{}
	reflectResources := reflect.ValueOf(resources)

	for i := 0; i < reflectResources.Len(); i++ {
		deleteID, err := getDeleteID(resType)
		if err != nil {
			return nil, err
		}

		deleteIDField, err := getField(deleteID, reflect.Indirect(reflectResources.Index(i)))
		if err != nil {
			return nil, errors.Wrapf(err, "Field with delete ID required for deleting resource")
		}

		tags, err := findTags(reflectResources.Index(i))
		if err != nil {
			logrus.WithError(err).Debug()
		}

		var creationTime *time.Time
		creationTimeField, err := findField(creationTimeFieldNames, reflect.Indirect(reflectResources.Index(i)))
		if err == nil {
			creationTimeCastTime, ok := creationTimeField.Interface().(*time.Time)
			if ok {
				creationTime = creationTimeCastTime
			} else {
				creationTimeCastString, ok := creationTimeField.Interface().(*string)
				if ok {
					parsedCreationTime, err := time.Parse("2006-01-02T15:04:05.000Z0700", *creationTimeCastString)
					if err == nil {
						creationTime = &parsedCreationTime
					}
				}
			}
		}

		deletableResources = append(deletableResources, &Resource{
			Type:    resType,
			ID:      deleteIDField.Elem().String(),
			Tags:    tags,
			Created: creationTime,
		})
	}

	return deletableResources, nil
}

func getField(name string, v reflect.Value) (reflect.Value, error) {
	field := v.FieldByName(name)

	if !field.IsValid() {
		return reflect.Value{}, errors.Errorf("Field not found: %s", name)
	}
	return field, nil
}

func findField(names []string, v reflect.Value) (reflect.Value, error) {
	for _, name := range names {
		field, err := getField(name, v)
		if err == nil {
			return field, nil
		}
	}
	return reflect.Value{}, errors.Errorf("Fields not found: %s", names)
}

// findTags finds findTags via reflection in the describe output.
func findTags(res reflect.Value) (map[string]string, error) {
	tags := map[string]string{}

	ts, err := findField(tagFieldNames, reflect.Indirect(res))
	if err != nil {
		return nil, errors.Wrap(err, "No tags found")
	}

	for i := 0; i < ts.Len(); i++ {
		key := reflect.Indirect(ts.Index(i)).FieldByName("Key").Elem()
		value := reflect.Indirect(ts.Index(i)).FieldByName("Value").Elem()
		tags[key.String()] = value.String()
	}

	return tags, nil
}

func (resources Resources) ToJson() (string, error) {
	b, err := json.Marshal(resources)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (resources Resources) ToYaml() (string, error) {
	b, err := yaml.Marshal(resources)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (resources Resources) ToString() (string, error) {
	printStat := ""
	for _, r := range resources {
		printStat += fmt.Sprintf("\tId:\t\t%s", r.ID)
		if r.Tags != nil {
			if len(r.Tags) > 0 {
				printStat += "\n\tTags:\t\t"
				for k, v := range r.Tags {
					printStat += fmt.Sprintf("[%s: %v] ", k, v)
				}
			}
		}
		printStat += "\n"
		if r.Created != nil {
			printStat += fmt.Sprintf("\tCreated:\t%s", r.Created)
			printStat += "\n"
		}
	}

	return printStat, nil
}
