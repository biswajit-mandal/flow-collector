# Query API Server

Query API Server (qAPI) is REST API Server built mainly to Store/Query flow messages.
Flow translator pushes ipfix and sflow data as sent by [vFlow](https://github.com/VerizonDigital/vflow) to Query API
Server. Once qAPI receives the flow data, it pushes to Database, currently qAPI stores the data in MongoDB.

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
$make build
```
It creates ```query-api``` binary inside ```api``` directory which provides Command Line Interface (CLI) utility, it can be invoked as below
```
$./query-api --help
NAME:
   query-api - A new cli application

USAGE:
   query-api [global options] command [command options] [arguments...]

VERSION:
   0.1

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config-file FILE  Load configuration from FILE (default: "/etc/query-api/query-api.conf")
   --help, -h          show help
   --version, -v       print the version```
```
To start the Query API Server:
```
$make run
```
# Configuration Parameters
The configuration file (query-api.conf) is an yml file with the below possible configurations
```
verbose: True
listen-port: 8080
use-database: "mongo"
mongo-ip: "127.0.0.1"
mongo-port: "27017"
#mongo-user: ""
#mongo-password: ""
log-file: "/var/log/query-api.log"
```

```verbose:``` Boolean, if Verbose mode is on or off (Default: False)

```listen-port:``` The port on which Query API Server should be listening on (Default: 8080)

```use-database``` Which database to use, currently only ```mongo``` is supported

```mongo-ip:``` IP Address where Mongo is running (Default: 127.0.0.1)

```mongo-port``` Mongo DB Port (Default: 27017)

```mongo-user``` User name to connect to Mongo DB

```mongo-password``` Password to connect to Mongo DB

```log-file:``` The file path where the log file should be created (Default: /var/log/query-api.log)

### REST APIs
qAPI provides 2 APIs
```
/create
/query
```
Both are POST request.
/create is used to create entry in Database
/query is used to retrieve the data from Database

#### Create entry
the POST data format for ```/create```
{"table_name": "<table/collection name>", "data":{}}
Sample IPFIX Flow data for pushing to Query API Server
```json
{"table_name":"ipfix_collection","data":{"AgentID":"10.84.30.149","Header":{"DomainID":524288,"ExportTime":1521144421,"Length":104,"SequenceNo":17463991,"Version":10},"DataSets":{"bgpDestinationAsNumber":64512,"bgpSourceAsNumber":64512,"destinationIPv4Address":"10.84.30.201","destinationIPv4PrefixLength":29,"destinationTransportPort":33930,"dot1qCustomerVlanId":0,"dot1qVlanId":0,"egressInterface":588,"flowEndMilliseconds":1521144417601,"flowEndReason":3,"flowStartMilliseconds":1521144417601,"fragmentIdentification":0,"icmpTypeCodeIPv4":0,"ingressInterface":556,"ipClassOfService":0,"ipNextHopIPv4Address":"10.84.30.141","maximumTTL":59,"minimumTTL":59,"octetDeltaCount":40,"packetDeltaCount":1,"protocolIdentifier":6,"sourceIPv4Address":"10.102.44.34","sourceIPv4PrefixLength":0,"sourceTransportPort":9000,"tcpControlBits":"0x14","vlanId":0},"Timestamp":1521139883428}}
```
### Query to retrieve data
The query is similar to SQL query with select, where and groupby clauses. Currently only where clause is implemented.

##### Where Clause
```time-series``` query
In time-series query, time range can be specified in below two ways:
Time Series Query can be specified with below 2 ways:

###### 'now-nx' format with start_time and end_time
Sample POST data for this as below:
```json
{"table_name": "ipfix_collection", "where":[{"start_time": "now-2m"}, {"end_time": "now"}]}
```
Will return last 2 minutes of data

So using start_time and end_time where user can specify time in ```now-nx``` format where ```n``` is any number and ```x``` can be any of the below
```d``` for day
```h``` for hour
```m``` for minute
```s``` for second

###### Using actual time range with actual keys
POST data for this as below
```json
{"table_name": "ipfix_collection", "where":[{"data.Timestamp": 1521139913413, "operator": ">="}, {"data.Timestamp": 1521139813413, "operator": "<="}]}
```
Here using exact keys as stored in DB along with operator key.
Operator keys can be ```>=``` or ```<=``` or ```>``` or ```<```
Operator key is optional, if not specified then assumed to be "=", no need to specify "=" as operator explicitly.

Where clause is generic, any field in the structure can be used and we can specify as many number of keys in where clause.

##### Select Clause
In select clause aggregate and non-aggregate keys can be specified, they are mutually exclusive.
The field can be at any level in the JSON structure.
Sample POST data as below:
```json
{"table_name": "ipfix_collection", "where":[{"start_time": "now-2m"}, {"end_time": "now"}], "select": ["data.Header"]}
```

##### GroupBy Clause
We can specify aggregate keys in the ```select``` clause along with ```groupby``` clause.
Only SUM is allowed to specify as aggregate key
Sample POST data with aggregate key as below:
```json
{"table_name": "ipfix_collection", "where":[{"start_time": "now-30m"}, {"end_time": "now"}], "select": ["SUM(data.DataSets.octetDeltaCount)"], "groupby": ["data.AgentID"]}
```

