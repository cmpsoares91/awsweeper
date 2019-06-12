package filters

import (
	"log"
	"regexp"

	"github.com/cmpsoares91/awsweeper/pkg/aws"
	"github.com/sirupsen/logrus"
)

func (f Filter) byTags(resources aws.IResources) (filteredResources aws.IResources, err error) {
	if f.Tags == nil || len(*f.Tags) == 0 {
		logrus.Debug("Filter byTags is nil. Ignoring")
		return resources, err
	}

	logrus.WithField("Tags", f.Tags).Debug("Filtering resources based on Tags")

	for _, tag := range *f.Tags {
		for _, r := range resources {
			r.EnsureLazyLoaded()
			allTagsMatched := true
			for tagKey, tagValueRegex := range tag {
				resourceTags := r.GetTags()
				if resourceTags != nil {
					if tagVal, ok := (*resourceTags)[tagKey]; !ok {
						allTagsMatched = false
					} else {
						if matched, err := regexp.MatchString(tagValueRegex, tagVal); err != nil {
							log.Fatal(err)
						} else if !matched {
							allTagsMatched = false
						}
					}
				}
			}

			if allTagsMatched {
				filteredResources = append(filteredResources, r)
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"Before Filtering": len(resources),
		"After Filtering":  len(filteredResources),
	}).Debug("Filtered By Tags")
	return filteredResources, err
}
