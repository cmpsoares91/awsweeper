package aws

import (
	"fmt"
	"os"

	tfConfig "github.com/hashicorp/terraform/config"
	tf "github.com/hashicorp/terraform/terraform"
	"github.com/sirupsen/logrus"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

func CreateProvider(awsConfig *Config) (*tf.ResourceProvider, error) {
	p := aws.Provider()

	cfg := map[string]interface{}{
		"region":      awsConfig.Region,
		"profile":     awsConfig.Profile,
		"max_retries": awsConfig.MaxRetries,
	}

	rc, err := tfConfig.NewRawConfig(cfg)
	if err != nil {
		return nil, err
	}

	conf := tf.NewResourceConfig(rc)

	warns, errs := p.Validate(conf)
	if len(warns) > 0 {
		logrus.Warn(warns)
	}
	if len(errs) > 0 {
		fmt.Printf("errors: %s\n", errs)
		os.Exit(1)
	}

	if err := p.Configure(conf); err != nil {
		return nil, err
	}

	return &p, nil
}
