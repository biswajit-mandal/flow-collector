/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    mgo_query.go
 * details: Deals with the Create/Query request handlers for Mongo
 *
 */
package dbhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	opts "github.com/Juniper/collector/query-api/options"
	req "github.com/Juniper/collector/query-api/request"
	res "github.com/Juniper/collector/query-api/response"
)

type opValStruct struct {
	operator interface{}
	value    interface{}
}

type M map[string]interface{}

func (mg *MongoDBHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/query", req.IsValidRequest(mg.HandleGetQuery))
	mux.HandleFunc("/create", req.IsValidRequest(mg.HandleCreateQuery))
}

func (mg *MongoDBHandler) HandleCreateQuery(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		result, statusCode, err := mg.handleCreate(w, r)
		if err != nil {
			res.RespondErr(w, r, statusCode, "Create Request failed: ", err)
			return
		}
		res.Respond(w, r, http.StatusCreated, &result)
	default:
		res.RespondHTTPErr(w, r, http.StatusNotFound)
	}
}

func (mg *MongoDBHandler) HandleGetQuery(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		results, statusCode, err := mg.HandleQuery(w, r)
		if err != nil {
			res.RespondErr(w, r, statusCode, "Query Request failed: ", err)
			return
		}
		res.Respond(w, r, statusCode, &results)
	default:
		res.RespondHTTPErr(w, r, http.StatusNotFound)
	}
}

func (mg *MongoDBHandler) handleCreate(w http.ResponseWriter, r *http.Request) (interface{}, int, error) {
	var (
		p          M
		err        error
		collection string
	)
	db := mg.mgoDB
	if err = req.DecodeBody(r, &p); err != nil {
		return nil, http.StatusBadRequest, err
	}
	for key, value := range p {
		switch key {
		case opts.TableName:
			collection = value.(string)
		}
	}
	c := db.C(collection)
	if err = c.Insert(p); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	var emptyStruct struct{}
	return emptyStruct, http.StatusCreated, nil
}

func getTimeByNowPrefix(timeStr string) (int64, error) {
	/* Supported options: now-1s, now-1m, now-1h, now-1d */
	var (
		err        error
		dayOffSet  int64
		timeOffSet string
		dur        time.Duration //int64
	)
	now := time.Now()
	if strings.HasSuffix(timeStr, "d") { //Day
		fmt.Sscanf(timeStr, "now-%dd", &dayOffSet)
		offSetStr := strconv.FormatInt(int64(dayOffSet*24), 10)
		dur, err = time.ParseDuration(offSetStr + "h")
		if err != nil {
			opts.Logger.Printf("Time %s should be in now-N[d/h/m/s] format: %v", timeStr, err)
			return 0, err
		}
	} else if timeStr == opts.NowPrefix {
		dur = 0
	} else {
		fmt.Sscanf(timeStr, "now-%s", &timeOffSet)
		dur, err = time.ParseDuration(timeOffSet)
		if err != nil {
			opts.Logger.Printf("Time %s should be in now-N[d/h/m/s] format: %v", timeStr, err)
			return 0, err
		}
	}
	if dur > 0 {
		dur = -dur
	}
	return now.Add(dur).UnixNano() / 1000000, nil //Milliseconds
}

func decodeQueryTime(startTime interface{}, endTime interface{}) (int64, int64, error) {
	var (
		err          error
		startTimeInt int64
		endTimeInt   int64
		endTimeStr   string
	)
	if startTime == nil {
		/* startTime must be reqd */
		return 0, 0, nil
	}
	startTimeStr := startTime.(string)
	if endTime == nil {
		endTimeStr = "now"
	} else {
		endTimeStr = endTime.(string)
	}
	if strings.HasPrefix(startTimeStr, opts.NowPrefix) {
		startTimeInt, err = getTimeByNowPrefix(startTimeStr)
		if err != nil {
			return 0, 0, err
		}
	}
	if strings.HasPrefix(endTimeStr, opts.NowPrefix) {
		endTimeInt, err = getTimeByNowPrefix(endTimeStr)
		if err != nil {
			return 0, 0, err
		}
	}
	if startTimeInt > endTimeInt {
		err = fmt.Errorf("startTime %s is greater than endTime %s", startTimeStr, endTimeStr)
	}
	return startTimeInt, endTimeInt, err
}

func (mg *MongoDBHandler) getAllRecords(collection string) ([]interface{}, error) {
	var (
		err     error
		count   int
		results []interface{}
	)
	db := mg.mgoDB
	coll := db.C(collection)
	q := coll.Find(nil).Sort(opts.TimeStampMapKey)
	err = q.All(&results)
	count, err = q.Count()
	opts.Logger.Println("Total Data Count as:", err, count)
	return results, err
}

func getMongoAggByOperator(operator interface{}) string {
	switch operator {
	case ">":
		return "$gt"
	case ">=":
		return "$gte"
	case "<":
		return "$lt"
	case "<=":
		return "$lte"
	default:
		return ""
	}
}

func getMongoAggByOperators(operator1 interface{}, operator2 interface{}) (string, string) {
	var (
		opAgg1 string
		opAgg2 string
	)
	opAgg1 = getMongoAggByOperator(operator1)
	opAgg2 = getMongoAggByOperator(operator2)
	return opAgg1, opAgg2
}

func getAbsTimeByOperator(timeInt int64, timeOp interface{}) int64 {
	switch timeOp {
	case "<":
		timeInt = timeInt - 1
	case ">":
		timeInt = timeInt + 1
	}
	return timeInt
}

func computeTimeSliceToQueryChunk(startTime int64, endTime int64) int64 {
	return (endTime - startTime) / (int64)(opts.MongoConnectionPoolLen)
}

func getTimeRangesForQueryChunk(startTime int64, startTimeOp interface{},
	endTime int64, endTimeOp interface{},
	needQuerySplit bool) []M {
	var (
		timeRanges            []M
		timeSliceToQueryChunk int64
	)
	startTime = getAbsTimeByOperator(startTime, startTimeOp)
	endTime = getAbsTimeByOperator(endTime, endTimeOp)
	if !needQuerySplit {
		timeRanges = append(timeRanges, M{"start_time": startTime, "end_time": endTime})
		return timeRanges
	}
	timeSliceToQueryChunk = computeTimeSliceToQueryChunk(startTime, endTime)
	for i := startTime; i < endTime; i = i + timeSliceToQueryChunk + 1 {
		startTimeInt := i
		endTimeInt := i + timeSliceToQueryChunk
		if endTimeInt > endTime {
			endTimeInt = endTime
		}
		timeRanges = append(timeRanges, M{"start_time": startTimeInt, "end_time": endTimeInt})
	}
	return timeRanges
}

func buildQueryByWhere(whereClauses map[string][]opValStruct,
	timeSortKeyPresent bool, needQuerySplit bool) ([][]interface{}, error) {
	var (
		ok                bool
		op1               interface{}
		op2               interface{}
		err               error
		opAgg1            string
		opAgg2            string
		startTime         interface{}
		endTime           interface{}
		whereQuery        M
		timeRanges        []M
		endTimeInt        int64
		resultQuerys      [][]interface{}
		startTimeInt      int64
		startTimeJsonInt  json.Number
		endTimeJsonInt    json.Number
		startTimeProvided bool
		timeKeyProvided   bool
		timeStampValues   []opValStruct
	)
	if whereClauses == nil {
		return nil, nil
	}
	whereQuery = make(M)
	for key, values := range whereClauses {
		op1 = values[0].operator
		op2 = ""
		if len(values) > 1 {
			op2 = values[1].operator
		}
		if opts.StartTime == key {
			startTime = values[0].value
			startTimeProvided = true
			continue
		}
		if opts.EndTime == key {
			endTime = values[0].value
			continue
		}
		/* Max 2 operator per key */
		opAgg1, opAgg2 = getMongoAggByOperators(op1, op2)
		if opts.TimeStampMapKey == key {
			timeKeyProvided = true
			timeStampValues = values
			continue
		}
		if opAgg1 == "" && opAgg2 == "" {
			whereQuery[key] = values[0].value
		} else if opAgg1 == "" {
			whereQuery[key] = M{opAgg2: values[1].value}
		} else if opAgg2 == "" {
			whereQuery[key] = M{opAgg1: values[0].value}
		} else {
			whereQuery[key] = M{opAgg1: values[0].value, opAgg2: values[1].value}
		}
	}
	if startTimeProvided && timeKeyProvided {
		err = fmt.Errorf("Both %s and %s provided.", opts.StartTime, opts.TimeStampMapKey)
		return nil, err
	}
	if startTime != nil {
		startTimeInt, endTimeInt, err = decodeQueryTime(startTime, endTime)
		if err != nil {
			return nil, err
		}
		op1 = ">="
		op2 = "<="
	}
	if len(timeStampValues) > 0 {
		if len(timeStampValues) != 2 {
			err = fmt.Errorf("Provide both startTime and endTime")
			return nil, err
		}
		timeOp1 := timeStampValues[0].operator
		timeOp2 := timeStampValues[1].operator
		startTimeErr := fmt.Errorf("Invalid startTime")
		endTimeErr := fmt.Errorf("Invalid endTime")
		if timeOp1 == ">" || timeOp1 == ">=" {
			startTimeJsonInt, ok = timeStampValues[0].value.(json.Number)
			if !ok {
				return nil, startTimeErr
			}
			endTimeJsonInt, ok = timeStampValues[1].value.(json.Number)
			if !ok {
				return nil, endTimeErr
			}
			op1 = timeOp1
		} else {
			startTimeJsonInt, ok = timeStampValues[1].value.(json.Number)
			if !ok {
				return nil, startTimeErr
			}
			endTimeJsonInt, ok = timeStampValues[0].value.(json.Number)
			if !ok {
				return nil, endTimeErr
			}
			op2 = timeOp2
		}
		startTimeInt, err = startTimeJsonInt.Int64()
		if !ok {
			return nil, err
		}
		endTimeInt, err = endTimeJsonInt.Int64()
		if !ok {
			return nil, err
		}
	}
	timeRanges = getTimeRangesForQueryChunk(startTimeInt, op1, endTimeInt, op2, needQuerySplit)
	timeRangesLen := len(timeRanges)
	for i := 0; i < timeRangesLen; i++ {
		tmpWhereQuery := make(M)
		var resultQuery []interface{}
		for key, _ := range whereQuery {
			tmpWhereQuery[key] = whereQuery[key]
		}
		tmpWhereQuery[opts.TimeStampMapKey] = M{"$gte": timeRanges[i]["start_time"],
			"$lte": timeRanges[i]["end_time"]}
		resultQuery = append(resultQuery, M{"$match": tmpWhereQuery})
		if !timeSortKeyPresent {
			resultQuery = append(resultQuery, M{"$sort": M{opts.TimeStampMapKey: 1}})
		}
		resultQuerys = append(resultQuerys, resultQuery)
	}
	return resultQuerys, nil
}

func isAggregateKey(key string) bool {
	aggKeys := []string{opts.SumPrefix}
	aggKeysCnt := len(aggKeys)
	for i := 0; i < aggKeysCnt; i++ {
		if strings.HasPrefix(key, aggKeys[i]) {
			return true
		}
	}
	return false
}

func getMongoAggKeyByUserKey(userKey string) (string, error) {
	if strings.HasPrefix(userKey, opts.SumPrefix) {
		return "$sum", nil
	}
	err := fmt.Errorf("Not supported aggregation key %s", userKey)
	return "", err
}

func getAggNonAggSelectFields(selectFields []interface{}) ([]string, []string, error) {
	var (
		err        error
		aggKeys    []string
		selectCnt  int
		nonAggKeys []string
	)
	if selectFields != nil {
		selectCnt = len(selectFields)
	}
	for i := 0; i < selectCnt; i++ {
		selectField := selectFields[i].(string)
		if isAggregateKey(selectField) {
			if !strings.HasSuffix(selectField, ")") {
				err = fmt.Errorf("Aggregate key %s is not valid", selectField)
				return nil, nil, err
			}
			aggKeys = append(aggKeys, selectField)
		} else {
			nonAggKeys = append(nonAggKeys, selectField)
		}
	}
	if len(aggKeys) > 0 && len(nonAggKeys) > 0 {
		err := fmt.Errorf("aggregate and non-aggregate keys are mutually exclusive")
		return nil, nil, err
	}
	return aggKeys, nonAggKeys, nil
}

func buildQueryByGroupBy(aggKeys []string, groupBy []interface{}) (interface{}, error) {
	var (
		err         error
		parseQuery  M
		mongoAggKey string
	)
	if groupBy == nil {
		return nil, nil
	}
	parseQuery = make(M)
	groupByCnt := len(groupBy)
	aggKeysLen := len(aggKeys)
	if groupByCnt == 0 && aggKeysLen == 0 {
		return nil, nil
	}
	groupByQ := make(M)
	for i := 0; i < groupByCnt; i++ {
		groupByVal := groupBy[i].(string)
		/* Mongo throws error if key contains ".", so replace with "_" */
		groupByKey := strings.Replace(groupByVal, ".", "_", -1)
		groupByQ[groupByKey] = "$" + groupByVal
	}
	resultQuery := make(M)
	for i := 0; i < aggKeysLen; i++ {
		aggActKey := aggKeys[i]
		aggVal := strings.TrimLeft(strings.TrimRight(aggActKey, ")"), opts.SumPrefix)
		aggKey := strings.Replace(aggActKey, ".", "_", -1)
		mongoAggKey, err = getMongoAggKeyByUserKey(aggActKey)
		if err != nil {
			return nil, err
		}
		parseQuery[aggKey] = M{mongoAggKey: "$" + aggVal}
	}
	resultQuery["$group"] = parseQuery
	if groupByCnt == 0 {
		parseQuery["_id"] = 0
	} else {
		parseQuery["_id"] = groupByQ
	}
	resultQuery["$group"] = parseQuery
	return resultQuery, nil
}

func buildQueryBySelect(selectKeys []string) (interface{}, error) {
	var (
		parseQ  M
		resultQ M
	)
	if selectKeys == nil {
		return nil, nil
	}
	selectKeysLen := len(selectKeys)
	if selectKeysLen == 0 {
		return nil, nil
	}
	parseQ = make(M)
	resultQ = make(M)
	for i := 0; i < selectKeysLen; i++ {
		selectKey := selectKeys[i]
		parseQ[selectKey] = 1
	}
	/* Skip _id which is auto generated, so no use to user */
	parseQ["_id"] = 0
	resultQ["$project"] = parseQ
	return resultQ, nil
}

func buildFinalQueriesByWhereQs(whereQs [][]interface{}, operations []interface{}) []interface{} {
	var (
		allOperations []interface{}
		whereQsLen    int
	)

	if whereQs == nil {
		allOperations = []interface{}{operations}
	} else {
		whereQsLen = len(whereQs)
		allOperations = make([]interface{}, whereQsLen)
	}
	opLen := len(operations)
	for i := 0; i < whereQsLen; i++ {
		deepCopiedOps := make([]interface{}, 1)
		deepCopiedOps = whereQs[i]
		if opLen > 0 {
			deepCopiedOps = append(deepCopiedOps, operations...)
		}
		allOperations[i] = deepCopiedOps
	}
	return allOperations
}

func (mg *MongoDBHandler) executeQuery(collection string, operations []interface{}) ([]interface{}, error) {
	var (
		err           error
		results       []interface{}
		allResults    []interface{}
		operationsCnt int
	)
	if len(operations) == 0 {
		/* User has asked for all the records, User must not issue this */
		return mg.getAllRecords(collection)
	}
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	operationsCnt = len(operations)
	for i := 0; i < operationsCnt; i++ {
		operation := operations[i]
		mgoDbIdx := i
		if i > opts.MongoConnectionPoolLen-1 {
			mgoDbIdx = int(math.Mod(float64(i-opts.MongoConnectionPoolLen), float64(opts.MongoConnectionPoolLen)))
		}
		g.Go(func() error {
			db := mg.mgoDB
			if opts.DoNeedQuerySplit {
				db = mg.mgoDBs[mgoDbIdx]
			}
			coll := db.C(collection)
			pipe := coll.Pipe(operation)
			pipe.AllowDiskUse()
			if opts.Verbose {
				opts.Logger.Println("Query sent to Mongo ", mgoDbIdx, operation)
			}
			err = pipe.All(&results)
			if err != nil {
				opts.Logger.Println("Mongo Query error:", err, operation)
				return err
			}
			if results != nil {
				allResults = append(allResults, results...)
			}
			return err
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}
	opts.Logger.Printf("Responded with %d count", len(allResults))
	return allResults, nil
}

func getSortOrderByKey(order interface{}) (int, error) {
	var (
		err      error
		orderVal int
	)

	switch order {
	case "asc":
		orderVal = 1
	case "desc":
		orderVal = -1
	default:
		err = fmt.Errorf("Order can be 'asc' or 'desc', provided %s", order)
	}
	return orderVal, err
}

func buildQueryByLimit(limit int64) (interface{}, error) {
	var (
		limitQ M
	)
	if limit == 0 {
		return nil, nil
	}
	limitQ = make(M)
	limitQ["$limit"] = limit
	return limitQ, nil
}

func buildQueryBySort(sortBys []interface{}) (interface{}, bool, error) {
	var (
		err                error
		sortBysLen         int
		sortQ              M
		timeSortKeyPresent bool
	)
	if sortBys == nil {
		return nil, false, nil
	}
	sortBysLen = len(sortBys)
	sortData := make(M)
	sortQ = make(M)
	for i := 0; i < sortBysLen; i++ {
		switch sortbys := sortBys[i].(type) {
		case map[string]interface{}:
			for sortKey, sortOrder := range sortbys {
				if sortKey == opts.TimeStampMapKey {
					timeSortKeyPresent = true
				}
				sortOrder, err = getSortOrderByKey(sortOrder)
				if err != nil {
					return nil, timeSortKeyPresent, err
				}
				sortData[sortKey] = sortOrder
			}
		}
	}
	sortQ["$sort"] = sortData
	return sortQ, timeSortKeyPresent, nil
}

func (mg *MongoDBHandler) buildQuery(collection string, selectFields []interface{},
	whereClauses []interface{}, groupBy []interface{}, sortBy []interface{},
	limit int64) ([]interface{}, error) {
	var (
		err                error
		sortQ              interface{}
		whereLen           int
		operations         []interface{}
		aggSelectKeys      []string
		allOperations      []interface{}
		needQuerySplit     bool
		nonAggSelectKeys   []string
		timeSortKeyPresent bool
	)

	if whereClauses != nil {
		whereLen = len(whereClauses)
	}
	var whereData = make(map[string][]opValStruct)
	for i := 0; i < whereLen; i++ {
		var (
			whereKey string
			whereVal interface{}
			opVal    interface{}
		)
		switch wheres := whereClauses[i].(type) {
		case map[string]interface{}:
			for key, value := range wheres {
				value = value.(interface{})
				if opts.OperatorStr == key {
					opVal = value
				} else {
					whereKey = key
					whereVal = value
				}
			}
		}
		whereData[whereKey] = append(whereData[whereKey], opValStruct{operator: opVal, value: whereVal})
	}
	aggSelectKeys, nonAggSelectKeys, err = getAggNonAggSelectFields(selectFields)
	if err != nil {
		return nil, err
	}
	needQuerySplit = (len(aggSelectKeys) == 0) && (opts.DoNeedQuerySplit)
	/* Process GroupBy Clause */
	groupByQ, err := buildQueryByGroupBy(aggSelectKeys, groupBy)
	if opts.Verbose {
		opts.Logger.Println("Done groupByClause processing:", groupByQ)
	}
	if err != nil {
		return nil, err
	}
	if groupByQ != nil {
		operations = append(operations, groupByQ)
	}
	/* Process Select Clause */
	selectQ, err := buildQueryBySelect(nonAggSelectKeys)
	if opts.Verbose {
		opts.Logger.Println("Done selectClause processing:", selectQ)
	}
	if err != nil {
		return nil, err
	}
	if selectQ != nil {
		operations = append(operations, selectQ)
	}
	/* Process Sort Clause */
	sortQ, timeSortKeyPresent, err = buildQueryBySort(sortBy)
	if opts.Verbose {
		opts.Logger.Println("Done sortClause processing:", sortQ)
	}
	if err != nil {
		return nil, err
	}
	if sortQ != nil {
		operations = append(operations, sortQ)
	}
	/* Process Limit */
	limitQ, err := buildQueryByLimit(limit)
	if opts.Verbose {
		opts.Logger.Println("Done limit processing:", limitQ)
	}
	if err != nil {
		return nil, err
	}
	if limitQ != nil {
		operations = append(operations, limitQ)
	}
	/* Process Where Clause */
	if limitQ != nil {
		needQuerySplit = false
	}
	whereQs, err := buildQueryByWhere(whereData, timeSortKeyPresent, needQuerySplit)
	if opts.Verbose {
		opts.Logger.Println("Done whereClause processing:", whereQs)
	}
	if err != nil {
		return nil, err
	}
	allOperations = buildFinalQueriesByWhereQs(whereQs, operations)
	return allOperations, nil
}

func (mg *MongoDBHandler) HandleQuery(w http.ResponseWriter, r *http.Request) ([]interface{}, int, error) {
	var (
		err           error
		results       []interface{}
		collection    string
		allOperations []interface{}
	)
	if opts.Verbose {
		opts.Logger.Println("Started the Query Processing")
	}
	collection, allOperations, err = mg.ReadQuery(w, r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	results, err = mg.executeQuery(collection, allOperations)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return results, http.StatusOK, nil
}

func (mg *MongoDBHandler) ReadQuery(w http.ResponseWriter, r *http.Request) (string, []interface{}, error) {
	var (
		ok            bool
		err           error
		limit         int64
		sortBy        []interface{}
		groupBy       []interface{}
		collection    string
		selectFields  []interface{}
		whereClause   []interface{}
		limitJsonInt  json.Number
		allOperations []interface{}
	)
	var p M
	if err := req.DecodeBodyNumber(r, &p); err != nil {
		return collection, nil, err
	}
	ok = true
	for key, value := range p {
		switch key {
		case opts.TableName:
			collection, ok = value.(string)
		case opts.Select:
			selectFields, ok = value.([]interface{})
		case opts.Where:
			whereClause, ok = value.([]interface{})
		case opts.GroupBy:
			groupBy, ok = value.([]interface{})
		case opts.SortBy:
			sortBy, ok = value.([]interface{})
		case opts.Limit:
			limitJsonInt, ok = value.(json.Number)
			if !ok {
				break
			}
			limit, err = limitJsonInt.Int64()
			if err != nil {
				return collection, nil, err
			}
		}
		if !ok {
			err = fmt.Errorf("Bad request in key '%s'", key)
			return collection, nil, err
		}
	}
	allOperations, err = mg.buildQuery(collection, selectFields, whereClause, groupBy, sortBy, limit)
	return collection, allOperations, err
}
