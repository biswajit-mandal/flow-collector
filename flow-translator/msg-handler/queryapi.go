/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    queryapi.go
 * details: Deals with the handling messages to send to Query API Server
 *
 */
package msghandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	opts "github.com/Juniper/collector/flow-translator/options"
)

// QueryAPI structure
type QueryAPI struct {
	netClient *http.Client
}

// QueryAPIMessage structure as the data needs to be pushed to Query API Server
type QueryAPIMessage struct {
	TableName string      `json:"table_name"`
	Data      interface{} `json:"data"`
}

func (qm *QueryAPI) setup() error {
	qm.netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	return nil
}

func (qm *QueryAPI) handleMessages(mhChan chan []byte) {
	var (
		msg []byte
	)
	for {
		select {
		case msg = <-mhChan:
			if opts.Verbose {
				opts.Logger.Println("Received Message on Query API Handler ", string(msg))
			}
			qm.pushDataToQueryAPIByTopic(msg)
		}
	}
}

func (qm *QueryAPI) serializeDataByTopic(msg []byte) ([]QueryAPIMessage, error) {
	if opts.KafkaTopic == opts.KafkaTopicVFlowIPFIX {
		return SerializeIPFIXMsgInQA(msg)
	} else if opts.KafkaTopic == opts.KafkaTopicVFlowSFlow {
		return SerializeSFlowMsgInQA(msg)
	} else {
		opts.Logger.Println("Not supported Topic:", opts.KafkaTopic)
	}
	return nil, nil
}

func (qm *QueryAPI) pushDataToQueryAPIByTopic(msg []byte) {
	augMsgs, err := qm.serializeDataByTopic(msg)
	if err != nil {
		opts.Logger.Println("QA -> data serialize error ", err)
		return
	}
	qm.pushDataToQueryAPI(augMsgs)
}

func (qm *QueryAPI) pushDataToQueryAPI(qmMsgs []QueryAPIMessage) error {
	var (
		reqUrl      string
		contentType string
	)
	contentType = "application/json"
	msgCnt := len(qmMsgs)
	reqUrl = fmt.Sprintf("http://%s:%s/create", opts.QueryApiIPAddress,
		opts.QueryApiPort)
	for idx := 0; idx < msgCnt; idx++ {
		qmMsg, err := json.Marshal(qmMsgs[idx])
		if err != nil {
			opts.Logger.Println("data json.Marshal() error ", err)
		}
		if opts.Verbose {
			opts.Logger.Println("Sending POST data to Query API Server", reqUrl, string(qmMsg))
		}
		response, err := qm.netClient.Post(reqUrl, contentType,
			strings.NewReader(string(qmMsg)))
		if err != nil {
			opts.Logger.Println("QueryAPI POST error ", err)
		} else {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				opts.Logger.Fatalln("Parse error response body ", err)
			}
			if opts.Verbose {
				opts.Logger.Println("Getting response from Query API ", string(body))
			}
		}
	}
	return nil
}
