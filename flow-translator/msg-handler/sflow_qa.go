/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    sflow_qa.go
 * details: sFlow packet handler for Query API Server
 *
 */
package msghandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	opts "github.com/Juniper/collector/flow-translator/options"
)

// SflowQAMessage sFlow message structure required for Data Manager
type SflowQAMessage struct {
	ExtSWData map[string]interface{} `json:"ExtSWData"`
	Header    map[string]interface{} `json:"Header"`
	Packet    map[string]interface{} `json:"Packet"`
	Sample    map[string]interface{} `json:"Sample"`
	Timestamp interface{}            `json:"Timestamp"`
}

func serializeSFlowQAData(msg *SflowQAMessage) ([]QueryAPIMessage, error) {
	var (
		err       error
		timeStamp int64
	)

	res := make([]QueryAPIMessage, 1)
	err = fmt.Errorf("Invalid timeStamp in sflow msg")
	for key, value := range msg.Header {
		if key == "Timestamp" {
			timeStampJsonInt, ok := value.(json.Number)
			if !ok {
				return nil, err
			}
			timeStamp, err = timeStampJsonInt.Int64()
			if err != nil {
				opts.Logger.Println("sFlow Msg TimStamp decode err:", err)
				return nil, err
			}
			timeStamp = timeStamp / 1000 // timestamp in milliseconds
			break
		}
	}
	res[0] = QueryAPIMessage{TableName: opts.SFLOWCollection,
		Data: SflowQAMessage{Header: msg.Header, ExtSWData: msg.ExtSWData,
			Packet: msg.Packet, Sample: msg.Sample, Timestamp: timeStamp}}
	return res, nil
}

func SerializeSFlowMsgInQA(msg []byte) ([]QueryAPIMessage, error) {
	d := json.NewDecoder(bytes.NewReader(msg))
	d.UseNumber()
	var sfMsg SflowQAMessage
	if err := d.Decode(&sfMsg); err != nil {
		opts.Logger.Println("sFlow message decode error:", err)
		return nil, err
	}
	qaMsgs, err := serializeSFlowQAData(&sfMsg)
	return qaMsgs, err
}
