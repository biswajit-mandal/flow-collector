package dbhandler

import (
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	opts "github.com/Juniper/collector/query-api/options"
	mgo "gopkg.in/mgo.v2"
)

/* Tests */
var (
	StrTestValidStartTimeEndTimeInMins   = "valid start_time & end_time in mins"
	StrTestInvalidStartTimeEndTimeInMins = "invalid start_time & end_time in mins"
	StrTestValidStartTimeEndTimeInMS     = "valid start_time & end_time in milliseconds"
	StrTestInvalidStartTimeEndTimeInMS   = "invalid start_time & end_time in milliseconds"
	StrTestValidSelect                   = "valid select clause"
	StrTestInvalidSelect                 = "invalid select clause"
	StrTestValidGroupby                  = "valid GroupBy Clase"
	StrTestInvalidGroupby                = "invalid GroupBy Clase"
	StrTestValidWhere                    = "valid where clause"
	StrTestInvalidWhere                  = "invalid where clause"
	StrTestValidSort                     = "valid sort clause"
	StrTestInvalidSort                   = "invalid sort clause"
	StrTestValidLimit                    = "valid limit value"
	StrTestInvalidLimit                  = "invalid limit value"
)

type fields struct {
	mgoSession *mgo.Session
	mgoDB      *mgo.Database
	mgoDBs     []*mgo.Database
}

type testStruct struct {
	name   string
	fields fields
	uri    string
	body   string
	method string
}

type fn func(testStruct, *testing.T, string, []interface{}, error)

func TestHandleQuery(t *testing.T) {
	testFnsMap := map[string]fn{
		StrTestValidStartTimeEndTimeInMins:   CheckValidStartTimeEndTimeInMins,
		StrTestInvalidStartTimeEndTimeInMins: CheckInvalidStartTimeEndTimeInMins,
		StrTestValidStartTimeEndTimeInMS:     CheckValidStartTimeEndTimeInMS,
		StrTestInvalidStartTimeEndTimeInMS:   CheckInvalidStartTimeEndTimeInMS,
		StrTestValidSelect:                   CheckValidSelect,
		StrTestInvalidSelect:                 CheckInvalidSelect,
		StrTestValidGroupby:                  CheckValidGroupby,
		StrTestInvalidGroupby:                CheckInvalidGroupby,
		StrTestValidWhere:                    CheckValidWhere,
		StrTestInvalidWhere:                  CheckInvalidWhere,
		StrTestValidSort:                     CheckValidSort,
		StrTestInvalidSort:                   CheckInvalidSort,
		StrTestValidLimit:                    CheckValidLimit,
		StrTestInvalidLimit:                  CheckInvalidLimit,
	}
	tests := []testStruct{
		{
			name:   StrTestValidStartTimeEndTimeInMins,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-1d"}, {"end_time": "now"}]}`,
		},
		{
			name:   StrTestInvalidStartTimeEndTimeInMins,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-dh"}, {"end_time": "now"}]}`,
		},
		{
			name:   StrTestValidStartTimeEndTimeInMS,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"data.Timestamp": 1522053648000, "operator": ">="}, {"data.Timestamp": 1522050048000, "operator": "<="}]}`,
		},
		{
			name:   StrTestInvalidStartTimeEndTimeInMS,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"data.Timestamp": 1522053648000, "operator": ">="}, {"data.Timestamp": "now", "operator": "<="}]}`,
		},
		{
			name:   StrTestValidSelect,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-1h"}, {"end_time": "now"}], "select": ["data.DataSets.octetDeltaCount"]}`,
		},
		{
			name:   StrTestInvalidSelect,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-1h"}, {"end_time": "now"}], "select": "data.DataSets.octetDeltaCount"}`,
		},
		{
			name:   StrTestValidGroupby,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection", "where":[{"start_time": "now-1d"}, {"end_time": "now"}], "select": ["SUM(data.DataSets.octetDeltaCount)"], "groupby": ["data.AgentID"]}`,
		},
		{
			name:   StrTestInvalidGroupby,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection", "where":[{"start_time": "now-1d"}, {"end_time": "now"}], "select": ["SUM(data.DataSets.octetDeltaCount)", "data.AgentID"], "groupby": ["data.AgentID"]}`,
		},
		{
			name:   StrTestValidWhere,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"data.Header.SequenceNo": 123456}], "select": ["data.Header"]}`,
		},
		{
			name:   StrTestInvalidWhere,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":{"data.Header.SequenceNo": 123456}, "select": ["data.Header"]}`,
		},
		{
			name:   StrTestValidSort,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-1h"}, {"end_time": "now"}], "sort": [{"data.Timestamp": "desc"}], "order": "desc"}`,
		},
		{
			name:   StrTestInvalidSort,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-1h"}, {"end_time": "now"}], "sort": "data.Timestamp"}`,
		},
		{
			name:   StrTestValidLimit,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-1h"}, {"end_time": "now"}], "limit": 10}`,
		},
		{
			name:   StrTestInvalidLimit,
			method: "POST",
			uri:    "/query",
			body:   `{"table_name": "ipfix_collection","where":[{"start_time": "now-1h"}, {"end_time": "now"}], "limit": "wrong"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := &MongoDBHandler{
				mgoSession: tt.fields.mgoSession,
				mgoDB:      tt.fields.mgoDB,
				mgoDBs:     tt.fields.mgoDBs,
			}
			req := httptest.NewRequest(tt.method, tt.uri, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			collection, operations, err := mg.ReadQuery(w, req)
			testFnsMap[tt.name](tt, t, collection, operations, err)
		})
	}
}

func getErrorStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func testError(name string, t *testing.T, expected interface{}, result interface{}) {
	if expected != result {
		t.Errorf("%s failed, expected '%v', got '%v'", name, expected, result)
	}
}
func CheckValidStartTimeEndTimeInMins(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	if err != nil {
		testError(tt.name, t, nil, err.Error())
		return
	}
}

func CheckInvalidStartTimeEndTimeInMins(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	expected := "time: invalid duration dh"
	testError(tt.name, t, expected, getErrorStr(err))
}

func CheckValidStartTimeEndTimeInMS(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	if err != nil {
		testError(tt.name, t, nil, err.Error())
		return
	}
}
func CheckInvalidStartTimeEndTimeInMS(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	expected := "Invalid endTime"
	testError(tt.name, t, expected, getErrorStr(err))
}
func CheckValidSelect(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	if err != nil {
		testError(tt.name, t, nil, err.Error())
		return
	}
}
func CheckInvalidSelect(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	expected := "Bad request in key 'select'"
	testError(tt.name, t, expected, getErrorStr(err))
}
func CheckValidGroupby(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	if err != nil {
		testError(tt.name, t, nil, err.Error())
		return
	}
}
func CheckInvalidGroupby(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	expected := "aggregate and non-aggregate keys are mutually exclusive"
	testError(tt.name, t, expected, getErrorStr(err))
}
func CheckValidWhere(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	if err != nil {
		testError(tt.name, t, nil, err.Error())
		return
	}
}
func CheckInvalidWhere(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	expected := "Bad request in key 'where'"
	testError(tt.name, t, expected, getErrorStr(err))
}
func CheckValidSort(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	if err != nil {
		testError(tt.name, t, nil, err.Error())
		return
	}
}
func CheckInvalidSort(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	expected := "Bad request in key 'sort'"
	testError(tt.name, t, expected, getErrorStr(err))
}
func CheckValidLimit(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	if err != nil {
		testError(tt.name, t, nil, err.Error())
		return
	}
}
func CheckInvalidLimit(tt testStruct, t *testing.T, collection string, operations []interface{}, err error) {
	expected := "Bad request in key 'limit'"
	testError(tt.name, t, expected, getErrorStr(err))
}
func setup(needQSplit bool, connPoolLen int) {
	opts.Verbose = true
	opts.Logger = log.New(os.Stderr, "[QE] ", log.Ldate|log.Ltime)
	opts.DoNeedQuerySplit = needQSplit
	opts.MongoConnectionPoolLen = connPoolLen
}

func shutdown(retCode int) {
	log.Println("Test Done!!!")
	os.Exit(retCode)
}

func TestMain(m *testing.M) {
	setup(false, 10)
	/* First test with single Query */
	log.Printf("Starting the test Cases with DoNeedQuerySplit %v and MongoConnectionPoolLen %d",
		opts.DoNeedQuerySplit, opts.MongoConnectionPoolLen)
	retCode := m.Run()
	/* Now test with split Query */
	setup(true, 10)
	log.Printf("\n\n\nStarting the test Cases with DoNeedQuerySplit %v and MongoConnectionPoolLen %d",
		opts.DoNeedQuerySplit, opts.MongoConnectionPoolLen)
	retCode = m.Run()
	shutdown(retCode)
}
