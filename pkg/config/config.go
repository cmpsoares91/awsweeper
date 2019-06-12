package config

import (
	"fmt"

	"github.com/cmpsoares91/awsweeper/pkg/aws"
	"github.com/cmpsoares91/awsweeper/pkg/filters"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
)

// AppFs is an abstraction of the file system to allow mocking in tests.
var AppFs = afero.NewOsFs()

// Config represents the content of a yaml file that is used as a contract to filter resources for deletion.
type Config struct {
	Options Options                              `yaml:",omitempty"`
	Filters map[aws.ResourceType]filters.Filters `yaml:",omitempty"`
}

type Options struct {
	DryRun           bool              `yaml:"dry-run,omitempty"`
	MaxRetries       int               `yaml:"max-retries,omitempty"`
	S3ForcePathStyle bool              `yaml:"s3-force-path-style,omitempty"`
	Regions          []string          `yaml:"regions"`
	RoleToAssume     string            `yaml:"role-to-assume,omitempty"`
	Extra            map[string]string `yaml:"extra,omitempty"`
}

// Load will read yaml config file and returns its value as config type
func Load(filename string) (*Config, error) {
	var cfg Config

	data, err := afero.ReadFile(AppFs, filename)
	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(data, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Options.Regions == nil {
		return nil, fmt.Errorf("At least one region is required in options")
	}

	return &cfg, nil
}
