package main

import (
	ogb "github.com/savageking-io/ogbcommon"
	"github.com/savageking-io/ogbrest/user_client"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "ogbrest"
	app.Version = AppVersion
	app.Description = "Smart backend service for smart game developers"
	app.Usage = "REST Microservice of NoErrorCode ecosystem"

	app.Authors = []cli.Author{
		{
			Name:  "savageking.io",
			Email: "i@savageking.io",
		},
		{
			Name:  "Mike Savochkin (crioto)",
			Email: "mike@crioto.com",
		},
	}

	app.Copyright = "2025 (c) savageking.io. All Rights Reserved"

	app.Commands = []cli.Command{
		{
			Name:  "serve",
			Usage: "Start REST",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "config",
					Usage:       "Configuration filepath",
					Value:       ConfigFilepath,
					Destination: &ConfigFilepath,
				},
				cli.StringFlag{
					Name:        "log",
					Usage:       "Specify logging level",
					Value:       "",
					Destination: &LogLevel,
				},
			},
			Action: Serve,
		},
	}

	_ = app.Run(os.Args)
}

func Serve(c *cli.Context) error {
	err := ogb.ReadYAMLConfig(ConfigFilepath, &AppConfig)
	if err != nil {
		return err
	}

	if LogLevel == "" && AppConfig.LogLevel != "" {
		LogLevel = AppConfig.LogLevel
	}
	if LogLevel == "" {
		LogLevel = "info"
	}

	err = ogb.SetLogLevel(LogLevel)
	if err != nil {
		return err
	}

	log.Infof("Configuration loaded from %s", ConfigFilepath)
	log.Infof("REST server will start on %s:%d", AppConfig.Rest.Hostname, AppConfig.Rest.Port)
	log.Infof("Configured %d services", len(AppConfig.Services))

	userClient := user_client.UserClient{}
	if err := userClient.Init(AppConfig.UserClient.Hostname, AppConfig.UserClient.Port); err != nil {
		return err
	}

	if err := userClient.Start(); err != nil {
		log.Errorf("Failed to start user client: %s", err.Error())
		userClient.Restart() // This will continuously attempt to reconnect
	}

	rest := REST{}
	err = rest.Init(&AppConfig.Rest, &userClient)
	if err != nil {
		return err
	}

	service := Service{}
	err = service.Init(&rest)
	if err != nil {
		return err
	}
	if err := service.Start(&rest); err != nil {
		log.Errorf("Failed to start service: %s", err.Error())
		return err
	}

	if err := rest.Start(); err != nil {
		log.Errorf("Failed to start REST: %s", err.Error())
	}
	return nil
}
