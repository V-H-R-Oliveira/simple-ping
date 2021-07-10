package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"log"
	"net"
	"syscall"
	"time"
)

func generateRequestID() uint16 {
	buffer := make([]byte, 2)

	if _, err := rand.Read(buffer); err != nil {
		log.Fatal(err)
	}

	return binary.BigEndian.Uint16(buffer)
}

func computeChecksum(packet []byte) uint16 {
	if (len(packet) & 1) == 1 {
		packet = append(packet, 0)
	}

	sum := 0

	for i := 0; i < len(packet); i += 4 {
		sum += int(binary.BigEndian.Uint32(packet[i : i+4]))
	}

	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	return uint16(^sum)
}

func closeSocket(fd int) {
	if err := syscall.Close(fd); err != nil {
		log.Fatal("Error on closing the socket:", err)
	}
}

func calculatePadding(dataSize int) int {
	alignedPacketLength := 0

	if dataSize < 32 {
		alignedPacketLength = 32 * (32 / dataSize)
	} else {
		alignedPacketLength = 32 * (dataSize / 32)
	}

	return alignedPacketLength - dataSize - 8 // icmp header without data has 8 bytes
}

func getTimestamp() *syscall.Timeval {
	timeval := syscall.Timeval{}

	if err := syscall.Gettimeofday(&timeval); err != nil {
		log.Fatal("Error on getting time of day:", err)
	}

	return &timeval
}

func ParseTimestamp(timestamp []byte) time.Time {
	return time.Unix(int64(binary.LittleEndian.Uint64(timestamp[:8])), int64(binary.LittleEndian.Uint64(timestamp[8:16])))
}

func readResponse(socketFd int, ch chan<- []byte, packetSize int) {
	response := make([]byte, packetSize)

	if _, _, err := syscall.Recvfrom(socketFd, response, syscall.MSG_WAITALL); err != nil {
		log.Println("Failed to receive a response due error:", err)
		ch <- response
		return
	}

	ch <- response
}

func parseIpHeader(payload []byte) (*IpProtoHeader, []byte) {
	return &IpProtoHeader{
		Version:        payload[0] >> 4,
		IHL:            payload[0] ^ (DEFAULT_VERSION << 4),
		TypeOfService:  payload[1],
		TotalLength:    binary.BigEndian.Uint16(payload[2:4]),
		Identifier:     binary.BigEndian.Uint16(payload[4:6]),
		Flags:          binary.BigEndian.Uint16(payload[6:8]) >> 13,
		FragmentOffset: binary.BigEndian.Uint16(payload[6:8]),
		TTL:            payload[8],
		Protocol:       payload[9],
		Checksum:       binary.BigEndian.Uint16(payload[10:12]),
		SrcAddr:        net.IPv4(payload[12], payload[13], payload[14], payload[15]),
		DstAddr:        net.IPv4(payload[16], payload[17], payload[18], payload[19]),
	}, payload[20:]
}

func parseIcmpReply(payload []byte) *IcmpProto {
	return &IcmpProto{
		Type:           payload[0],
		Code:           payload[1],
		Checksum:       binary.BigEndian.Uint16(payload[2:4]),
		Identifier:     binary.BigEndian.Uint16(payload[4:6]),
		SequenceNumber: binary.BigEndian.Uint16(payload[6:8]),
		Data:           payload[8:],
	}
}
