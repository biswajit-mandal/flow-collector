/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    ipfix_qa.go
 * details: IPFIX packet handler for Query API Server
 *
 */
package msghandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	opts "github.com/Juniper/collector/flow-translator/options"
)

// Message represents IPFIX message
type IPFIXQAMessage struct {
	AgentID   string                   `json:"AgentID"`
	Header    map[string]interface{}   `json:"Header"`
	DataSets  []map[string]interface{} `json:"DataSets"`
	Timestamp interface{}              `json:"Timestamp"`
}

// Query API, we split the data into length of DataSets.
type IPFIXAugmentedQAMessage struct {
	AgentID   string                 `json:"AgentID"`
	Header    map[string]interface{} `json:"Header"`
	DataSets  map[string]interface{} `json:"DataSets"`
	Timestamp interface{}            `json:"Timestamp"`
}

func serializeIPFIXQAData(msg *IPFIXQAMessage) ([]QueryAPIMessage, error) {
	var (
		err       error
		timeStamp int64
	)
	res := make([]QueryAPIMessage, len(msg.DataSets))
	err = fmt.Errorf("Invalid timeStamp in ipfix msg")
	timeStampJsonInt, ok := msg.Timestamp.(json.Number)
	if !ok {
		return nil, err
	}
	timeStamp, err = timeStampJsonInt.Int64()
	if err != nil {
		opts.Logger.Println("ipfix Msg TimStamp decode err:", err)
		return nil, err
	}
	timeStamp = timeStamp / 1000
	for i, dataSet := range msg.DataSets {
		res[i] = QueryAPIMessage{TableName: opts.IPFIXCollection,
			Data: IPFIXAugmentedQAMessage{Header: msg.Header, DataSets: dataSet,
				AgentID:   msg.AgentID,
				Timestamp: timeStamp}}
	}
	return res, nil
}

func SerializeIPFIXMsgInQA(msg []byte) ([]QueryAPIMessage, error) {
	d := json.NewDecoder(bytes.NewReader(msg))
	d.UseNumber()
	var ipfixMsg IPFIXQAMessage
	if err := d.Decode(&ipfixMsg); err != nil {
		opts.Logger.Println("IPFIX message decode error:", err)
		return nil, err
	}
	qaMsgs, err := serializeIPFIXQAData(&ipfixMsg)
	return qaMsgs, err
}
