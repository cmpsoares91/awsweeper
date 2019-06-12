package filters

import (
	"github.com/cmpsoares91/awsweeper/pkg/aws"
	"github.com/sirupsen/logrus"
)

func (f Filter) byNot(resources aws.IResources) (filteredResources aws.IResources, err error) {
	if f.Not == nil {
		logrus.Debug("Filter byNot is nil. Ignoring")
		return resources, err
	}

	logrus.WithField("Not:", f.Not).Debug("Filtering resources based on Not")
	matchedResources, err := f.Not.Apply(resources)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"Number of Matched Resources": len(matchedResources),
	}).Debug("'Not' filter discovered resources that should not be included")
	matchedResourcesMap := make(map[string]bool)
	for _, mr := range matchedResources {
		matchedResourcesMap[mr.GetID()] = true
	}

	for _, r := range resources {
		if _, ok := matchedResourcesMap[r.GetID()]; !ok {
			filteredResources = append(filteredResources, r)
		}
	}

	logrus.WithFields(logrus.Fields{
		"Before Filtering": len(resources),
		"After Filtering":  len(filteredResources),
	}).Debug("Filtered By Not")
	return filteredResources, err
}
