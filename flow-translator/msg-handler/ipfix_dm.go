/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    ipfix_dm.go
 * details: IPFIX packet handler for Data Manager
 *
 */
package msghandler

import (
	"bytes"
	"encoding/json"
	opts "github.com/Juniper/collector/flow-translator/options"
)

// Message represents IPFIX message
type IPFIXDMMessage struct {
	AgentID   string                   `json:"AgentID"`
	Header    map[string]interface{}   `json:"Header"`
	DataSets  []map[string]interface{} `json:"DataSets"`
	Timestamp interface{}              `json:"Timestamp"`
	RoomKey   string                   `json:"roomKey"`
}

// Data Manager, we split the data into length of DataSets.
type IPFIXAugmentedDMMessage struct {
	AgentID   string                 `json:"AgentID"`
	Header    map[string]interface{} `json:"Header"`
	DataSets  map[string]interface{} `json:"DataSets"`
	Timestamp interface{}            `json:"Timestamp"`
	RoomKey   string                 `json:"roomKey"`
}

func serializeIPFIXData(msg *IPFIXDMMessage) []DMMessage {
	var (
		emptyData []DMMessage
	)
	res := make([]DMMessage, len(msg.DataSets))
	opts.Logger.Printf("getting msg.Timestamp as %v %T:", msg.Timestamp,
		msg.Timestamp)
	timeStamp, err := msg.Timestamp.(json.Number).Int64()
	if err != nil {
		opts.Logger.Println("TimStamp err:", err)
		return emptyData
	}
	timeStamp = timeStamp / 1000
	for i, dataSet := range msg.DataSets {
		res[i] = DMMessage{CollectionName: opts.IPFIXCollection,
			Data: IPFIXAugmentedDMMessage{Header: msg.Header, DataSets: dataSet,
				AgentID:   msg.AgentID,
				RoomKey:   msg.AgentID,
				Timestamp: timeStamp}}
		var emptyTM struct{}
		res[i].TailwindManager = &emptyTM
	}
	return res
}

func SerializeIPFIXMsgInDM(msg []byte) ([]DMMessage, error) {
	d := json.NewDecoder(bytes.NewReader(msg))
	d.UseNumber()
	var ipfixMsg IPFIXDMMessage
	if err := d.Decode(&ipfixMsg); err != nil {
		opts.Logger.Println("sFlow message decode error:", err)
		return nil, err
	}
	dmMsgs := serializeIPFIXData(&ipfixMsg)
	return dmMsgs, nil
}
