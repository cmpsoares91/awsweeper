package main

import (
	"fmt"
	"time"

	amazon "github.com/aws/aws-sdk-go/aws"
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

	awsConf := &amazon.Config{
		Region:     amazon.String("ap-southeast-1"),
		MaxRetries: amazon.Int(1),
	}

	provider, err := aws.CreateProvider(awsConf)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create aws provider")
	}

	client, err := aws.NewClient(awsConf)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create aws API client")
	}

	timeShift, _ := time.ParseDuration("24h")
	wiper := wipe.Wiper{
		DryRun:    true,
		Client:    client,
		Provider:  provider,
		Filters:   config,
		TimeShift: &timeShift,
	}

	resources, warnings, err := wiper.Run()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to wipe resources")
	}

	if len(warnings) > 0 {
		logrus.WithField("Warnings:", warnings).Warn("Unable to perform as expected because of these warnings")
	}

	fmt.Println(resources.ToJson())
}
