/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    sflow_qa_test.go
 * details: Deals with the Unit Test cases for the exported functions as defined in sflow_qa.go
 *
 */
package msghandler

import (
	"testing"
)

type sflowArgs struct {
	msg []byte
}

type sflowTestStruct struct {
	name    string
	args    sflowArgs
	want    []QueryAPIMessage
	wantErr bool
}

func TestSerializeSFlowMsgInQA(t *testing.T) {
	tests := []sflowTestStruct{
		{
			name: StrTestValidSFlowMessage,
			args: sflowArgs{
				msg: MockData[StrTestValidSFlowMessage],
			},
			wantErr: false,
		},
		{
			name: StrTestInvalidSFlowMessage,
			args: sflowArgs{
				msg: MockData[StrTestInvalidSFlowMessage],
			},
			wantErr: true,
		},
		{
			name: StrTestInvalidTSSFlowMessage,
			args: sflowArgs{
				msg: MockData[StrTestInvalidTSSFlowMessage],
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		testFnsMap := map[string]func(*testing.T, sflowTestStruct, []QueryAPIMessage, error){
			StrTestValidSFlowMessage:     CheckValidSFlowMessage,
			StrTestInvalidSFlowMessage:   CheckInvalidSFlowMessage,
			StrTestInvalidTSSFlowMessage: CheckInvalidTSSFlowMessage,
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := SerializeSFlowMsgInQA(tt.args.msg)
			testFnsMap[tt.name](t, tt, got, err)
		})
	}
}

func CheckValidSFlowMessage(t *testing.T, tt sflowTestStruct, result []QueryAPIMessage, err error) {
	if (err != nil) != tt.wantErr {
		VerifyError(tt.name, t, nil, err.Error())
	}
}

func CheckInvalidSFlowMessage(t *testing.T, tt sflowTestStruct, result []QueryAPIMessage, err error) {
	expected := "invalid character 'i' looking for beginning of value"
	if (err == nil) != tt.wantErr {
		VerifyError(tt.name, t, expected, err.Error())
	}
}

func CheckInvalidTSSFlowMessage(t *testing.T, tt sflowTestStruct, result []QueryAPIMessage, err error) {
	expected := "Invalid timeStamp in sflow msg"
	if (err == nil) != tt.wantErr {
		VerifyError(tt.name, t, expected, err.Error())
	}
}
