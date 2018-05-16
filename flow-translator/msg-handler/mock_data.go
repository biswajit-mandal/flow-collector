/*
 * Copyright (c) 2018 Juniper Networks, Inc. All rights reserved.
 *
 * file:    mock_data.go
 * details: Contains the mock data for the Unit Tests for msghandler package
 *
 */
package msghandler

import ()

/* Tests */
var (
	StrTestValidIPFIXMessage     = "Valid ipfix message"
	StrTestInvalidIPFIXMessage   = "Invalid ipfix message"
	StrTestInvalidTSIPFIXMessage = "Invalid timestamp in ipfix message"
	StrTestValidSFlowMessage     = "Valid sflow message"
	StrTestInvalidSFlowMessage   = "Invalid sflow message"
	StrTestInvalidTSSFlowMessage = "Invalid timestamp in sflow message"
)

var MockData map[string][]byte

func InitMockData() {
	MockData = map[string][]byte{
		/* IPFIX Data */
		StrTestValidIPFIXMessage: []byte(`{"AgentID":"10.84.30.149","Timestamp":1522035620663678,"Header":{"Version":10,"Length":104,"ExportTime":1522040162,"SequenceNo":19127360,"DomainID":524288},"DataSets":[{"sourceIPv4Address":"10.84.29.30","destinationIPv4Address":"10.84.30.218","ipClassOfService":0,"protocolIdentifier":6,"sourceTransportPort":55246,"destinationTransportPort":8780,"icmpTypeCodeIPv4":0,"ingressInterface":556,"vlanId":0,"sourceIPv4PrefixLength":0,"destinationIPv4PrefixLength":29,"bgpSourceAsNumber":64512,"bgpDestinationAsNumber":64512,"ipNextHopIPv4Address":"10.84.30.165","tcpControlBits":"0x11","egressInterface":573,"octetDeltaCount":52,"packetDeltaCount":1,"minimumTTL":60,"maximumTTL":60,"flowStartMilliseconds":1522040157115,"flowEndMilliseconds":1522040157115,"flowEndReason":3,"dot1qVlanId":0,"dot1qCustomerVlanId":0,"fragmentIdentification":0}]}`),
		/* Invalid Message, Timestamp is wrong type */
		StrTestInvalidTSIPFIXMessage: []byte(`{"AgentID":"10.84.30.149","Timestamp":"abcd","Header":{"Version":10,"Length":104,"ExportTime":1522040162,"SequenceNo":19127360,"DomainID":524288},"DataSets":[{"sourceIPv4Address":"10.84.29.30","destinationIPv4Address":"10.84.30.218","ipClassOfService":0,"protocolIdentifier":6,"sourceTransportPort":55246,"destinationTransportPort":8780,"icmpTypeCodeIPv4":0,"ingressInterface":556,"vlanId":0,"sourceIPv4PrefixLength":0,"destinationIPv4PrefixLength":29,"bgpSourceAsNumber":64512,"bgpDestinationAsNumber":64512,"ipNextHopIPv4Address":"10.84.30.165","tcpControlBits":"0x11","egressInterface":573,"octetDeltaCount":52,"packetDeltaCount":1,"minimumTTL":60,"maximumTTL":60,"flowStartMilliseconds":1522040157115,"flowEndMilliseconds":1522040157115,"flowEndReason":3,"dot1qVlanId":0,"dot1qCustomerVlanId":0,"fragmentIdentification":0}]}`),
		StrTestInvalidIPFIXMessage:   []byte(`invalid ipfix message`),
		/* SFLOW Data */
		StrTestValidSFlowMessage:     []byte(`{"Header":{"Version":5,"IPVersion":1,"AgentSubID":16,"SequenceNo":55739,"SysUpTime":1104544871,"SamplesNo":1,"Timestamp":1522108407531878,"IPAddress":"10.84.30.141"},"ExtSWData":{"SrcVlan":0,"SrcPriority":0,"DstVlan":0,"DstPriority":0},"Sample":{"SequenceNo":132547,"SourceID":0,"SamplingRate":2560,"SamplePool":1249241330,"Drops":0,"Input":505,"Output":0,"RecordsNo":2},"Packet":{"L2":{"SrcMAC":"00:25:90:94:b4:e6","DstMAC":"54:e0:32:88:73:81","Vlan":0,"EtherType":2048},"L3":{"Version":4,"TOS":0,"TotalLen":52,"ID":37626,"Flags":0,"FragOff":0,"TTL":64,"Protocol":6,"Checksum":25392,"Src":"10.84.30.201","Dst":"172.29.111.95"},"L4":{"SrcPort":9092,"DstPort":54510,"DataOffset":8,"Reserved":0,"Flags":16}}}`),
		StrTestInvalidTSSFlowMessage: []byte(`{"Header":{"Version":5,"IPVersion":1,"AgentSubID":16,"SequenceNo":55739,"SysUpTime":1104544871,"SamplesNo":1,"Timestamp":"wrong","IPAddress":"10.84.30.141"},"ExtSWData":{"SrcVlan":0,"SrcPriority":0,"DstVlan":0,"DstPriority":0},"Sample":{"SequenceNo":132547,"SourceID":0,"SamplingRate":2560,"SamplePool":1249241330,"Drops":0,"Input":505,"Output":0,"RecordsNo":2},"Packet":{"L2":{"SrcMAC":"00:25:90:94:b4:e6","DstMAC":"54:e0:32:88:73:81","Vlan":0,"EtherType":2048},"L3":{"Version":4,"TOS":0,"TotalLen":52,"ID":37626,"Flags":0,"FragOff":0,"TTL":64,"Protocol":6,"Checksum":25392,"Src":"10.84.30.201","Dst":"172.29.111.95"},"L4":{"SrcPort":9092,"DstPort":54510,"DataOffset":8,"Reserved":0,"Flags":16}}}`),
		StrTestInvalidSFlowMessage:   []byte(`invalid sflow message`),
	}
}
