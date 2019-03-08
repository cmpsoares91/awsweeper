package config

import (
	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
)

// AppFs is an abstraction of the file system to allow mocking in tests.
var AppFs = afero.NewOsFs()

// Config represents the content of a yaml file that is used as a contract to filter resources for deletion.
type Config = aws.Filters

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

	return &cfg, nil
}
