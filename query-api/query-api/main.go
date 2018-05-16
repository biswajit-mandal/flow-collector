/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    main.go
 * details: Entry point for Query API Server, the binary creates Command
 *			Line Interface (CLI) utility to run the application.
 *
 */
package main

import (
	"net/http"
	"os"

	dbhandler "github.com/Juniper/collector/query-api/db-handler"
	opts "github.com/Juniper/collector/query-api/options"

	"github.com/urfave/cli"
)

var (
	version string
)

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-file",
			Value: "/etc/query-api/query-api.conf",
			Usage: "Load configuration from `FILE`",
		},
	}
	app.Action = RunQueryAPIServer
	err := app.Run(os.Args)
	if err != nil {
		opts.Logger.Fatalf("Application error: %s", err)
	}
}

func RunQueryAPIServer(c *cli.Context) error {
	opts.ParseArgs(c)
	StartQueryAPIServer()
	return nil
}

func StartQueryAPIServer() {
	mux := http.NewServeMux()
	registerDBHandlers(mux)
	opts.Logger.Println("Starting Web Server on :", opts.ListenPort)
	http.ListenAndServe(":"+opts.ListenPort, mux)
}

func registerDBHandlers(mux *http.ServeMux) {
	dbHandlersLen := len(opts.DataBaseList)
	for i := 0; i < dbHandlersLen; i++ {
		go func(i int) {
			dbH := dbhandler.NewDBHandler(opts.DataBaseList[i])
			if err := dbH.Run(mux); err != nil {
				opts.Logger.Fatalf("dbHandler %v run error %v ", opts.DataBaseList[i], err)
			}
		}(i)
	}
}
