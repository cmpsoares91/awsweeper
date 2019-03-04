package main

import (
	"fmt"

	"github.com/iflix/awsweeper/pkg/aws"
	"github.com/iflix/awsweeper/pkg/config"
	"github.com/iflix/awsweeper/pkg/wipe"
	"github.com/sirupsen/logrus"
)

func main() {
	yamlFilePath := "config.yaml"
	config, err := config.Load(yamlFilePath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open config file")
	}

	awsConf := &aws.Config{
		Region:     "ap-southeast-1",
		Profile:    "",
		MaxRetries: 1,
	}

	provider, err := aws.CreateProvider(awsConf)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create aws provider")
	}

	client, err := aws.NewClient(awsConf)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create aws API client")
	}

	wiper := wipe.Wiper{
		DryRun:   true,
		Client:   client,
		Provider: provider,
		Filters:  &config,
	}

	resources, err := wiper.Run()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to wipe resources")
	}

	fmt.Println(resources.ToJson())
}
