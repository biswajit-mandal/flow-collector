/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    main.go
 * details: Entry point for the ipfix-translator, the binary creates Command
 *          Line Interface (CLI) utility to run the application.
 */
package main

import (
	"os"

	kc "github.com/Juniper/collector/flow-translator/kafka-consumer"
	opts "github.com/Juniper/collector/flow-translator/options"
	"github.com/urfave/cli"
)

var version string

func handleKafkaConsumer(c *cli.Context) error {
	opts.ParseArgs(c)
	kc.KafkaConsumer()
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "Kafka Consumer CLI"
	app.Version = version
	app.Commands = []cli.Command{
		{
			Name:  "kafka-consumer",
			Usage: "Kafka Consumer",
			Flags: []cli.Flag{
				cli.StringFlag{Name: opts.MHConfigFileStr, Value: opts.MHConfigFile,
					Usage: "The config file"},
			},
			Action: handleKafkaConsumer,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		opts.Logger.Fatalf("Application error: %s", err)
	}
}
