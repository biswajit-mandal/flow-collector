/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    sflow_dm.go
 * details: sFlow packet handler for Data Manager
 *
 */
package msghandler

import (
	"bytes"
	"encoding/json"
	opts "github.com/Juniper/collector/flow-translator/options"
)

// SflowDMMessage sFlow message structure required for Data Manager
type SflowDMMessage struct {
	ExtSWData map[string]interface{} `json:"ExtSWData"`
	Header    map[string]interface{} `json:"Header"`
	Packet    map[string]interface{} `json:"Packet"`
	Sample    map[string]interface{} `json:"Sample"`
	Timestamp interface{}            `json:"Timestamp"`
	RoomKey   string                 `json:"roomKey"`
}

func serializeSFlowData(msg *SflowDMMessage) []DMMessage {
	var (
		err            error
		timeStamp      int64
		timeStampFound bool
		roomKey        string
		roomKeyFound   bool
		emptyData      []DMMessage
	)

	res := make([]DMMessage, 1)

	for key, value := range msg.Header {
		if key == "Timestamp" {
			timeStamp, err = value.(json.Number).Int64()
			if err != nil {
				opts.Logger.Println("TimStamp err:", err)
				return emptyData
			}
			timeStamp = timeStamp / 1000 // timestamp in milliseconds
			timeStampFound = true
		}
		if key == "IPAddress" {
			roomKey = value.(string)
			roomKeyFound = true
		}
		if (timeStampFound) && (roomKeyFound) {
			break
		}
	}
	opts.Logger.Println("getting timeStamp as:", timeStamp)
	res[0] = DMMessage{CollectionName: opts.SFLOWCollection,
		Data: SflowDMMessage{Header: msg.Header, ExtSWData: msg.ExtSWData,
			Packet: msg.Packet, Sample: msg.Sample,
			RoomKey: roomKey, Timestamp: timeStamp}}
	var emptyTM struct{}
	res[0].TailwindManager = &emptyTM
	return res
}

func SerializeSFlowMsgInDM(msg []byte) ([]DMMessage, error) {
	d := json.NewDecoder(bytes.NewReader(msg))
	d.UseNumber()
	var sfMsg SflowDMMessage
	if err := d.Decode(&sfMsg); err != nil {
		opts.Logger.Println("sFlow message decode error:", err)
		return nil, err
	}
	dmMsgs := serializeSFlowData(&sfMsg)
	return dmMsgs, nil
}
