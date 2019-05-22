package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/sirupsen/logrus"
)

type ElasticSearchDomainApi struct {
	api *elasticsearchservice.ElasticsearchService
}

func (a *ElasticSearchDomainApi) getType() ResourceType {
	return "elasticsearch_domain"
}

func (a *ElasticSearchDomainApi) getPriority() int64 {
	return -1
}

func (a *ElasticSearchDomainApi) new(s *session.Session, cfg *aws.Config) {
	a.api = elasticsearchservice.New(s, cfg)
}

func (a *ElasticSearchDomainApi) list() (resources IResources, err error) {
	listDomainNamesOutput, err := a.api.ListDomainNames(&elasticsearchservice.ListDomainNamesInput{})
	if err != nil {
		return resources, err
	}

	for _, domainName := range listDomainNamesOutput.DomainNames {
		r := &ElasticSearchDomain{
			ID:           domainName.DomainName,
			Name:         domainName.DomainName,
			Tags:         make(Tags),
			ResourceType: a.getType(),
			api:          a.api,
		}

		resources = append(resources, r)
	}

	return resources, nil
}

// ElasticSearchDomain ...
type ElasticSearchDomain Resource

// Delete ...
func (r *ElasticSearchDomain) Delete() error {
	logrus.WithField("ElasticSearchDomain", *r.Name).Info("Deleting ElasticSearchDomain")
	api := r.api.(*elasticsearchservice.ElasticsearchService)
	result, err := api.DeleteElasticsearchDomain(&elasticsearchservice.DeleteElasticsearchDomainInput{DomainName: r.ID})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Result": result.String(),
		"Name":   *r.Name,
	}).Info("ElasticSearchDomain deleted")
	return nil
}

// String ...
func (r *ElasticSearchDomain) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *ElasticSearchDomain) GetID() string {
	if r.ID != nil {
		return *r.ID
	}

	return ""
}

// GetName ...
func (r *ElasticSearchDomain) GetName() string {
	if r.Name != nil {
		return *r.Name
	}

	return ""
}

// GetTags ...
func (r *ElasticSearchDomain) GetTags() *Tags { return &r.Tags }

// GetCreationDate ...
func (r *ElasticSearchDomain) GetCreationDate() *time.Time { return r.CreationDate }

// EnsureLazyLoaded ...
func (r *ElasticSearchDomain) EnsureLazyLoaded() {
	if !r.lazyLoadPerformed {
		logrus.WithField("resource", r).Debug("Performing a lazyload on a elastic search domain")
		api := r.api.(*elasticsearchservice.ElasticsearchService)
		domainDesc, err := api.DescribeElasticsearchDomain(&elasticsearchservice.DescribeElasticsearchDomainInput{DomainName: r.ID})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load ESD description")
		}

		configOutput, err := api.DescribeElasticsearchDomainConfig(&elasticsearchservice.DescribeElasticsearchDomainConfigInput{DomainName: r.ID})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load ESD config")
		}

		r.CreationDate = configOutput.DomainConfig.AdvancedOptions.Status.CreationDate

		tagsOutput, err := api.ListTags(&elasticsearchservice.ListTagsInput{ARN: domainDesc.DomainStatus.ARN})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load ESD tags")
		}

		if tagsOutput.TagList != nil {
			for _, tag := range tagsOutput.TagList {
				r.Tags[*tag.Key] = *tag.Value
			}
		}

		r.lazyLoadPerformed = true
	}
}
