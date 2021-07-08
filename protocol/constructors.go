package protocol

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"net"
	"syscall"
)

func NewIPHeader(dstAddr net.IP) *IpProtoHeader {
	return &IpProtoHeader{
		Version:        DEFAULT_VERSION,
		IHL:            MIN_IP_HEADER_LENGTH,
		TypeOfService:  0,
		TotalLength:    0,
		Identifier:     0,
		Flags:          NO_FRAGMENT,
		FragmentOffset: 0,
		TTL:            MAX_TTL,
		Protocol:       ICMP_PROTO,
		Checksum:       0,
		SrcAddr:        nil,
		DstAddr:        dstAddr.To4(),
	}
}

func (ipProto *IpProtoHeader) ToBytes() []byte {
	buffer := bytes.Buffer{}

	binary.Write(&buffer, binary.BigEndian, (ipProto.Version<<4)+ipProto.IHL)
	binary.Write(&buffer, binary.BigEndian, ipProto.TypeOfService)
	binary.Write(&buffer, binary.BigEndian, ipProto.TotalLength)
	binary.Write(&buffer, binary.BigEndian, ipProto.Identifier)
	binary.Write(&buffer, binary.BigEndian, (ipProto.Flags<<13)+ipProto.FragmentOffset)
	binary.Write(&buffer, binary.BigEndian, ipProto.TTL)
	binary.Write(&buffer, binary.BigEndian, ipProto.Protocol)
	binary.Write(&buffer, binary.BigEndian, ipProto.Checksum)
	binary.Write(&buffer, binary.BigEndian, uint32(0))
	binary.Write(&buffer, binary.BigEndian, ipProto.DstAddr)

	return buffer.Bytes()
}

func NewIcmpProto() *IcmpProto {
	packet := &IcmpProto{
		Type:           ECHO_REQUEST,
		Code:           0,
		Checksum:       0,
		Identifier:     generateRequestID(),
		SequenceNumber: 0,
		Data:           []byte{},
	}

	return packet
}

func (packet *IcmpProto) SetTimestamp() {
	timeval := getTimestamp()

	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, timeval)

	unpaddedData := buffer.Bytes()
	padding := calculatePadding(len(unpaddedData))

	for i := 0; i < padding; i++ {
		binary.Write(&buffer, binary.BigEndian, uint8(0))
	}

	packet.Data = buffer.Bytes()
}

func (packet *IcmpProto) SetSequenceNumber(sequenceNumber uint16) {
	packet.SequenceNumber = sequenceNumber
}

func (packet *IcmpProto) ComputeChecksum() {
	packet.ClearChecksum()
	packet.Checksum = computeChecksum(packet.ToBytes())
}

func (packet *IcmpProto) ClearChecksum() {
	packet.Checksum = 0
}

func (packet *IcmpProto) ToBytes() []byte {
	buffer := bytes.Buffer{}

	binary.Write(&buffer, binary.BigEndian, packet.Type)
	binary.Write(&buffer, binary.BigEndian, packet.Code)
	binary.Write(&buffer, binary.BigEndian, packet.Checksum)
	binary.Write(&buffer, binary.BigEndian, packet.Identifier)
	binary.Write(&buffer, binary.BigEndian, packet.SequenceNumber)
	binary.Write(&buffer, binary.BigEndian, packet.Data)

	return buffer.Bytes()
}

func NewIcmpRequest(dst net.IP) *IcmpRequest {
	return &IcmpRequest{
		IpHeader: NewIPHeader(dst),
		Request:  NewIcmpProto(),
	}
}

func (icmpReq *IcmpRequest) ToBytes() []byte {
	buffer := []byte{}

	buffer = append(buffer, icmpReq.IpHeader.ToBytes()...)
	buffer = append(buffer, icmpReq.Request.ToBytes()...)

	return buffer
}

func (icmp *IcmpRequest) SendRequest(dstAddr [4]byte) {
	dst := syscall.SockaddrInet4{
		Port: 0,
		Addr: dstAddr,
	}

	socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)

	if err != nil {
		log.Fatal("Error on creating the socket:", err)
	}

	defer closeSocket(socketFd)

	if err := syscall.Sendto(socketFd, icmp.ToBytes(), 0, &dst); err != nil {
		log.Fatal("Error on sending packet:", err)
	}
}

func GetReply(ctx context.Context, packetSize int) *IcmpReply {
	ch := make(chan []byte)
	go readResponse(ch, packetSize)

	select {
	case <-ctx.Done():
		log.Printf("Default timeout of %d seconds exceeded.\n", DEFAULT_TIMEOUT)
		return nil
	case response := <-ch:
		ipHeader, icmpPayload := parseIpHeader(response)
		icmpReply := parseIcmpReply(icmpPayload)

		return &IcmpReply{
			IpHeader: ipHeader,
			Reply:    icmpReply,
		}
	}
}
