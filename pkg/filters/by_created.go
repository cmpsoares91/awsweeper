package filters

import (
	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/sirupsen/logrus"
)

func (f Filter) byCreated(resources aws.IResources) (filteredResources aws.IResources, err error) {
	if f.Created == nil {
		logrus.Debug("Filter byCreated is nil. Ignoring")
		return resources, err
	}

	logrus.WithField("Created:", f.Created).Debug("Filtering resources based on CreatedDate")
	for _, r := range resources {
		createdAfter := true
		createdBefore := true
		r.EnsureLazyLoaded()
		creationDate := r.GetCreationDate()

		if creationDate != nil {
			if f.Created.After != nil {
				createdAfter = creationDate.Unix() > f.Created.After.Unix()
			}

			if f.Created.Before != nil {
				createdBefore = creationDate.Unix() < f.Created.Before.Unix()
			}
		} else {
			logrus.WithField("Resource", r).Warn("Ignoring 'Created' filtering because resources does not have creation date")
		}

		if createdAfter && createdBefore {
			filteredResources = append(filteredResources, r)
		}
	}

	logrus.WithFields(logrus.Fields{
		"Number of Resources": len(resources),
		"After filtering":     len(filteredResources),
	}).Debug("Filtered By Created")
	return filteredResources, err
}
