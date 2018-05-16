/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    main.go
 * details: Deals with the handling messages to send to Data Manager
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

// DataManager structure
type DataManager struct {
	netClient *http.Client
}

// DMMessage structure as the data needs to be pushed to DM
type DMMessage struct {
	CollectionName  string      `json:"collection_name"`
	Data            interface{} `json:"data"`
	TailwindManager interface{} `json:"tailwind_manager"`
}

func (dm *DataManager) setup() error {
	dm.netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	return nil
}

func (dm *DataManager) handleMessages(mhChan chan []byte) {
	var (
		msg []byte
	)
	for {
		select {
		case msg = <-mhChan:
			if opts.Verbose {
				opts.Logger.Println("Received Message on DM Handler ", string(msg))
			}
			dm.pushDataToDataManagerByTopic(msg)
		}
	}
}

func (dm *DataManager) serializeDataByTopic(msg []byte) ([]DMMessage, error) {
	if opts.KafkaTopic == opts.KafkaTopicVFlowIPFIX {
		return SerializeIPFIXMsgInDM(msg)
	} else if opts.KafkaTopic == opts.KafkaTopicVFlowSFlow {
		return SerializeSFlowMsgInDM(msg)
	} else {
		opts.Logger.Println("Not supported Topic:", opts.KafkaTopic)
	}
	return nil, nil
}

func (dm *DataManager) pushDataToDataManagerByTopic(msg []byte) {
	augMsgs, err := dm.serializeDataByTopic(msg)
	if err != nil {
		opts.Logger.Println("DM -> data serialize error ", err)
		return
	}
	dm.pushDataToDataManager(augMsgs)
}

func (dm *DataManager) pushDataToDataManager(dmMsgs []DMMessage) error {
	var (
		reqUrl      string
		contentType string
	)
	contentType = "application/json"
	msgCnt := len(dmMsgs)
	reqUrl = fmt.Sprintf("http://%s:%s/version/2.0/post_event", opts.DataMgrIpAddress, opts.DataMgrPort)
	for idx := 0; idx < msgCnt; idx++ {
		dmMsg, err := json.Marshal(dmMsgs[idx])
		if err != nil {
			opts.Logger.Println("data json.Marshal() error ", err)
		}
		if opts.Verbose {
			opts.Logger.Println("Sending POST data to DM ", reqUrl, string(dmMsg))
		}
		response, err := dm.netClient.Post(reqUrl, contentType, strings.NewReader(string(dmMsg)))
		if err != nil {
			opts.Logger.Println("DataManager POST error ", err)
		} else {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				opts.Logger.Fatalln("Parse error response body ", err)
			}
			if opts.Verbose {
				opts.Logger.Println("Getting response from DM ", string(body))
			}
		}
	}
	return nil
}
