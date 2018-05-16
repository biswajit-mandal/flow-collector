# IPFIX Translator

Flow Translator is translator piece of code for IPFIX & sFlow messages as received on Kafka Bus.
Messages from different devices are received by [vFlow](https://github.com/VerizonDigital/vflow) module and it then sends them on Kafka bus.
Once Flow Translator receives the messages on Kafka bus, it parses the data and pushes to different databases based on configuration. Currently it pushes the data only to Appformix Data Manager (DM).

### Build pre-requisites
- Install [git](https://www.atlassian.com/git/tutorials/install-git)
- Install [go](https://golang.org/doc/install)
- Install [dep](https://github.com/golang/dep)

### Get Dependencies
```
make deps
```
This will download all the dependencies.

### Build
```
make build
```
It creates ```flow-translator``` binary which provides Command Line Interface (CLI) utility, it can be invoked as below
```
$ ./flow-translator --help
NAME:
   Kafka Consumer CLI - A new cli application

USAGE:
   flow-translator [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     kafka-consumer  Kafka Consumer
     help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

To run the Flow Translator:
```
./flow-translator kafka-consumer --config-file /etc/flow-translator/flow-translator.conf
```
# Configuration Parameters
The configuration file (flow-translator.conf) is an yml file with the below possible configurations
```
verbose: True
log-file: "/var/log/flow-translator.log"
kafka-broker-list: "127.0.0.1:9092"
kafka-topic: "vflow.ipfix"

query-api-ip: "127.0.0.1"
query-api-port: "8080"
```

```verbose:``` Boolean, if Verbose mode is on or off

```log-file:``` The file path where the log file should be created

```kafka-broker-list:``` Kafka Broker List

```kafka-topic:``` Kafka Topic, For IPFIX,```vflow.ipfix``` and for sFlow,```vflow.sflow```

```query-api-ip``` IP of the Query API Server

```query-api-port``` Port of the Query API Server

