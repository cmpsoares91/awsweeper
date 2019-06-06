package main

import (
	"fmt"

	"github.com/iflix/awsweeper/pkg/config"
	"github.com/iflix/awsweeper/pkg/wipe"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	yamlFilePath := "config.yaml"
	config, err := config.Load(yamlFilePath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open config file")
	}

	wiper := wipe.Wiper{
		Config: config,
	}

	resources, warnings, err := wiper.Run()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to wipe resources")
	}

	if len(warnings) > 0 {
		logrus.WithField("Warnings:", warnings).Warn("Unable to perform as expected because of these warnings")
	}

	fmt.Println(resources)
}
