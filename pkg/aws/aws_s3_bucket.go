package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

type S3BucketAPI struct {
	api *s3.S3
}

func (a *S3BucketAPI) getType() ResourceType {
	return "aws_s3_bucket"
}

func (a *S3BucketAPI) getPriority() int64 {
	return 9750
}

func (a *S3BucketAPI) new(s *session.Session, cfg *aws.Config) {
	a.api = s3.New(s, cfg)
}

func (a *S3BucketAPI) list() (resources IResources, err error) {
	buckets, err := a.api.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	for _, bucket := range buckets.Buckets {
		r := &S3Bucket{
			Name:         bucket.Name,
			ID:           bucket.Name,
			Tags:         make(Tags),
			CreationDate: bucket.CreationDate,
			ResourceType: a.getType(),
			api:          a.api,
		}

		bucketLocationOutput, err := a.api.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: bucket.Name})

		if err != nil {
			return nil, err
		}

		var bucketLocation string
		if bucketLocationOutput.LocationConstraint == nil {
			bucketLocation = "us-east-1"
		} else {
			bucketLocation = *bucketLocationOutput.LocationConstraint
		}

		if a.api.SigningRegion == bucketLocation {
			resources = append(resources, r)
		} else {
			logrus.WithFields(logrus.Fields{
				"Signing Region":  a.api.SigningRegion,
				"Bucket Location": bucketLocation,
				"Bucket":          *bucket.Name,
			}).Debug("Bucket is not in the current region. Skipping")
		}
	}

	return resources, err
}

// S3Bucket ...
type S3Bucket Resource

// Delete ...
func (r *S3Bucket) Delete() error {
	logrus.WithField("Bucket", *r.Name).Info("Deleting a Bucket")
	api := r.api.(*s3.S3)

	if dbop, err := api.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{Bucket: r.ID}); err != nil {
		return err
	} else {
		logrus.WithFields(logrus.Fields{
			"Result": dbop.String(),
			"Bucket": *r.Name,
		}).Info("Bucket policy deleted")

		result, err := api.DeleteBucket(&s3.DeleteBucketInput{Bucket: r.ID})

		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"Result": result.String(),
			"Bucket": *r.Name,
		}).Info("Bucket deleted")
	}
	return nil
}

func (r *S3Bucket) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// GetID ...
func (r *S3Bucket) GetID() *string { return r.ID }

// GetName ...
func (r *S3Bucket) GetName() *string { return r.Name }

// GetTags ...
func (r *S3Bucket) GetTags() *Tags {
	return &r.Tags
}

// GetCreationDate ...
func (r *S3Bucket) GetCreationDate() *time.Time { return r.CreationDate }

func (r *S3Bucket) EnsureLazyLoaded() {
	if !r.lazyLoadPerformed {
		logrus.WithField("resource", r).Debug("Performing a lazyload on a bucket")
		api := r.api.(*s3.S3)

		if taggingOutput, err := api.GetBucketTagging(&s3.GetBucketTaggingInput{Bucket: r.ID}); err == nil {
			if taggingOutput.TagSet != nil {
				for _, tag := range taggingOutput.TagSet {
					r.Tags[*tag.Key] = *tag.Value
				}
			}
		} else {
			logrus.WithError(err).Fatal("Failed to load Tags")
		}

		r.lazyLoadPerformed = true
	}
}
