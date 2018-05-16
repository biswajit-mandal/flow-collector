/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    ipfix_qa_test.go
 * details: Deals with the Unit Test cases for the exported functions as defined in ipfix_qa.go
 *
 */
package msghandler

import (
	"testing"
)

type ipfixArgs struct {
	msg []byte
}

type IPFIXtestStruct struct {
	name    string
	args    ipfixArgs
	want    []QueryAPIMessage
	wantErr bool
}

func TestSerializeIPFIXMsgInQA(t *testing.T) {
	tests := []IPFIXtestStruct{
		{
			name: StrTestValidIPFIXMessage,
			args: ipfixArgs{
				msg: MockData[StrTestValidIPFIXMessage],
			},
			wantErr: false,
		},
		{
			name: StrTestInvalidIPFIXMessage,
			args: ipfixArgs{
				msg: MockData[StrTestInvalidIPFIXMessage],
			},
			wantErr: true,
		},
		{
			name: StrTestInvalidTSIPFIXMessage,
			args: ipfixArgs{
				msg: MockData[StrTestInvalidTSIPFIXMessage],
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		testFnsMap := map[string]func(*testing.T, IPFIXtestStruct, []QueryAPIMessage, error){
			StrTestValidIPFIXMessage:     CheckValidIPFIXMessage,
			StrTestInvalidIPFIXMessage:   CheckInvalidIPFIXMessage,
			StrTestInvalidTSIPFIXMessage: CheckInvalidTSIPFIXMessage,
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := SerializeIPFIXMsgInQA(tt.args.msg)
			testFnsMap[tt.name](t, tt, got, err)
		})
	}
}

func CheckValidIPFIXMessage(t *testing.T, tt IPFIXtestStruct, result []QueryAPIMessage, err error) {
	if (err != nil) != tt.wantErr {
		VerifyError(tt.name, t, nil, err.Error())
	}
}

func CheckInvalidIPFIXMessage(t *testing.T, tt IPFIXtestStruct, result []QueryAPIMessage, err error) {
	expected := "invalid character 'i' looking for beginning of value"
	if (err == nil) != tt.wantErr {
		VerifyError(tt.name, t, expected, err.Error())
	}
}

func CheckInvalidTSIPFIXMessage(t *testing.T, tt IPFIXtestStruct, result []QueryAPIMessage, err error) {
	expected := "Invalid timeStamp in ipfix msg"
	if (err == nil) != tt.wantErr {
		VerifyError(tt.name, t, expected, err.Error())
	}
}
