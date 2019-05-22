package filters

import (
	"time"

	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/sirupsen/logrus"
)

func (f Filter) byAge(resources aws.IResources) (filteredResources aws.IResources, err error) {
	if f.Age == nil {
		logrus.Debug("Filter byAge is nil. Ignoring")
		return resources, err
	}

	logrus.WithField("Age:", f.Age).Debug("Filtering resources based on Age")
	now := time.Now()

	for _, r := range resources {
		createdAfter := true
		createdBefore := true
		r.EnsureLazyLoaded()
		creationDate := r.GetCreationDate()
		if creationDate != nil {
			if f.Age.YoungerThan != nil {
				createdAfter = creationDate.Unix() > now.Add(-*f.Age.YoungerThan).Unix()
			}

			if f.Age.OlderThan != nil {
				createdBefore = creationDate.Unix() < now.Add(-*f.Age.OlderThan).Unix()
			}
		}

		if createdAfter && createdBefore {
			filteredResources = append(filteredResources, r)
		}
	}

	logrus.WithFields(logrus.Fields{
		"Before Filtering": len(resources),
		"After Filtering":  len(filteredResources),
	}).Debug("Filtered By Age")
	return filteredResources, err
}
