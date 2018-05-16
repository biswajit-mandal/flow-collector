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
	Verbose           bool   `yaml:"verbose" env:"IPFIX_TRANSLATOR_LOG_ENABLE"`
	KafkaBrokerList   string `yaml:"kafka-broker-list" env:"KAFKA_BROKER_LIST"`
	KafkaTopic        string `yaml:"kafka-topic" env:"KAFKA_TOPIC"`
	DataMgrIpAddress  string `yaml:"data-manager-ip" env:"DATA_MANAGER_IP_ADDRESS"`
	DataMgrPort       string `yaml:"data-manager-port" env:"DATA_MANAGER_PORT"`
	QueryApiIPAddress string `yaml:"query-api-ip" env:"QUERY_API_IP_ADDRESS"`
	QueryApiPort      string `yaml:"query-api-port" env:"QUERY_API_PORT"`
	LogFile           string `yaml:"log-file" env:"IPFIX_LOG_FILE"`
	SendToDM          bool   `yaml:"sendto-data-manager" env:"SENDTO_DATA_MANAGER"`
	SendToQA          bool   `yaml:"sendto-query-api" env:"SENDTO_QUERY_API"`
}

var (
	Verbose              = false
	KafkaBrokerList      = "127.0.0.1:9092"
	KafkaTopicVFlowIPFIX = "vflow.ipfix"
	KafkaTopicVFlowSFlow = "vflow.sflow"
	DataMgrIpAddress     = "127.0.0.1"
	DataMgrPort          = "9000"
	QueryApiIPAddress    = "127.0.0.1"
	QueryApiPort         = "8080"
	LogFile              = "/var/log/flow-translator.log"
	SendToDM             = false
	SendToQA             = true

	StrDataManager     = "data-manager"
	StrQueryAPI        = "query-api"
	StrKafkaConGroupID = "ipfixConsGrpID"
	MHConfigFileStr    = "config-file"
	MHConfigFile       = "/etc/flow-translator/flow-translator.conf"
	KafkaTopic         = KafkaTopicVFlowIPFIX
	IPFIXCollection    = "ipfix_collection"
	SFLOWCollection    = "sflow_collection"
)

var Logger *log.Logger

// ParseArgs parses the arguments as passed from Command Line Interface (CLI)
func ParseArgs(c *cli.Context) error {
	MHConfigFile = c.String(MHConfigFileStr)
	b, err := ioutil.ReadFile(MHConfigFile)
	if err != nil {
		log.Fatalln("config file read error ", err)
	}
	config := ConfigOptions{
		Verbose:          Verbose,
		KafkaBrokerList:  KafkaBrokerList,
		DataMgrIpAddress: DataMgrIpAddress,
		DataMgrPort:      DataMgrPort,
		LogFile:          LogFile,
		KafkaTopic:       KafkaTopic,
		SendToDM:         SendToDM,
		SendToQA:         SendToQA,
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		log.Fatalf("Config file %v parse error: %v", MHConfigFile, err)
	}
	KafkaBrokerList = config.KafkaBrokerList
	DataMgrIpAddress = config.DataMgrIpAddress
	DataMgrPort = config.DataMgrPort
	Verbose = config.Verbose
	LogFile = config.LogFile
	KafkaTopic = config.KafkaTopic
	SendToDM = config.SendToDM
	SendToQA = config.SendToQA
	if LogFile != "" {
		Logger = log.New(os.Stderr, "[jFlow] ", log.Ldate|log.Ltime)
		f, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Logger.Println(err)
		} else {
			Logger.SetOutput(f)
		}
	}
	return nil
}
