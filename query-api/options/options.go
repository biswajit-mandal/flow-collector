/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    options.go
 * details: Deals with the all the global and configuration parameters
 *
 */
package options

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// ConfigOptions configuration options
type ConfigOptions struct {
	Verbose         bool   `yaml:"verbose" env:"QUERY_API_LOG_ENABLE"`
	ListenPort      string `yaml:"listen-port" env:"QUERY_API_LISTEN_PORT"`
	UseDatabase     string `yaml:"use-database" env:"QUERY_API_USE_DATABASE"`
	MongoIP         string `yaml:"mongo-ip" env:"QUERY_API_MONGO_IP"`
	MongoPort       string `yaml:"mongo-port" env:"QUERY_API_MONGO_PORT"`
	MongoUserName   string `yaml:"mongo-user" env:"QUERY_API_MONGO_USER"`
	MongoUserPasswd string `yaml:"mongo-password" env:"QUERY_API_MONGO_PASSWD"`
	LogFile         string `yaml:"log-file" env:"QUERY_API_LOG_FILE"`
}

var (
	Verbose         = false
	ListenPort      = "8080"
	MongoIP         = "127.0.0.1"
	MongoPort       = "9092"
	MongoUserName   = ""
	MongoUserPasswd = ""
	ConfigFile      = "/etc/query-api/query-api.conf"
	LogFile         = "/var/log/query-api.log"
	UseDatabase     = UseDatabaseMongo

	/* DB Selection */
	UseDatabaseMongo = "mongo"

	/* Collections */
	IPFIXCollection = "ipfix_collection"
	SFlowCollection = "sflow_collection"

	/* Create Query */
	DBFlows   = "flows_db"
	TableName = "table_name"
	Data      = "data"

	/* Get Query */
	Limit                  = "limit"
	SortBy                 = "sort"
	Select                 = "select"
	Where                  = "where"
	GroupBy                = "groupby"
	StartTime              = "start_time"
	EndTime                = "end_time"
	SumPrefix              = "SUM("
	NowPrefix              = "now"
	OperatorStr            = "operator"
	DataBaseList           = []string{UseDatabaseMongo}
	ConfigFileStr          = "config-file"
	TimeStampMapKey        = "data.Timestamp"
	DoNeedQuerySplit       = false
	MongoConnectionPoolLen = 10
)

var Logger *log.Logger

func ParseArgs(c *cli.Context) error {
	ConfigFile = c.String(ConfigFileStr)
	b, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		log.Fatalln("Config file read error ", err)
	}
	config := ConfigOptions{
		Verbose:         Verbose,
		ListenPort:      ListenPort,
		UseDatabase:     UseDatabase,
		MongoIP:         MongoIP,
		MongoPort:       MongoPort,
		MongoUserName:   MongoUserName,
		MongoUserPasswd: MongoUserPasswd,
		LogFile:         LogFile,
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		log.Fatalf("Config file %v parse error: %v", ConfigFile, err)
	}
	Verbose = config.Verbose
	ListenPort = config.ListenPort
	UseDatabase = config.UseDatabase
	MongoIP = config.MongoIP
	MongoPort = config.MongoPort
	MongoUserName = config.MongoUserName
	MongoUserPasswd = config.MongoUserPasswd
	LogFile = config.LogFile
	if LogFile != "" {
		Logger = log.New(os.Stderr, "[QE] ", log.Ldate|log.Ltime)
		f, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Logger.Println(err)
		} else {
			Logger.SetOutput(f)
		}
	}
	return nil
}
