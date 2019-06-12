package filters

import (
	"regexp"

	"github.com/cmpsoares91/awsweeper/pkg/aws"
	"github.com/sirupsen/logrus"
)

func (f Filter) byIDs(resources aws.IResources) (filteredResources aws.IResources, err error) {
	if f.IDs == nil || len(*f.IDs) == 0 {
		logrus.Debug("Filter byIDs is nil. Ignoring")
		return resources, err
	}

	logrus.WithField("IDs:", f.IDs).Debug("Filtering resources based on IDs")
	for _, idFilter := range *f.IDs {
		for _, r := range resources {
			if ok, err := regexp.MatchString(idFilter, r.GetID()); ok {
				if err != nil {
					return nil, err
				}

				filteredResources = append(filteredResources, r)
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"Before Filtering": len(resources),
		"After Filtering":  len(filteredResources),
	}).Debug("Filtered By ID")
	return filteredResources, err
}
