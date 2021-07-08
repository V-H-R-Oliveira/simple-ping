package protocol

import "net"

type IpProtoHeader struct {
	Version        uint8
	IHL            uint8
	TypeOfService  uint8
	TotalLength    uint16
	Identifier     uint16
	Flags          uint16
	FragmentOffset uint16
	TTL            uint8
	Protocol       uint8
	Checksum       uint16
	SrcAddr        net.IP
	DstAddr        net.IP
}

type IcmpProto struct {
	Type           uint8
	Code           uint8
	Checksum       uint16
	Identifier     uint16
	SequenceNumber uint16
	Data           []byte
}

type IcmpRequest struct {
	IpHeader *IpProtoHeader
	Request  *IcmpProto
}

type IcmpReply struct {
	IpHeader *IpProtoHeader
	Reply    *IcmpProto
}

const (
	DEFAULT_VERSION      = 4
	MIN_IP_HEADER_LENGTH = 5
	ICMP_PROTO           = 1
	ECHO_REQUEST         = 8
	MAX_RESPONSE_LENGTH  = 65535
	NO_FRAGMENT          = 2
	MAX_TTL              = 255
	DEFAULT_TIMEOUT      = 10
)
