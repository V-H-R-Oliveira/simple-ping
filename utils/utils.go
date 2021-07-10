package utils

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/V-H-R-Oliveira/dns-client/protocol"
	"github.com/V-H-R-Oliveira/dns-client/utils"
	icmpUtils "github.com/V-H-R-Oliveira/simple-ping/protocol"
)

func IpToByteSlice(ip string) [4]byte {
	splittedIp := strings.Split(ip, ".")
	addr := [4]byte{}

	for i, octet := range splittedIp {
		parsedOctec, err := strconv.ParseUint(octet, 10, 8)

		if err != nil {
			log.Fatalf("Error on parsing octet %v: %s\n", octet, err.Error())
		}

		addr[i] = byte(parsedOctec)
	}

	return addr
}

func ResolveDomain(domain string) net.IP {
	domain = strings.ToLower(domain)

	if domain == "localhost" {
		return net.ParseIP("127.0.0.1")
	}

	socket, err := utils.CreateUDPDNSSocket()

	if err != nil {
		log.Fatal("Error on creating dns socket:", err)
	}

	defer socket.Close()

	query := protocol.NewDNSQuery(domain, protocol.A)
	query.SendRequest(socket)

	rawResponse := protocol.GetResponse(socket)
	_, response := protocol.ParseDNSResponse(rawResponse, false)

	rawIp := response.Answers[0].Data
	resolvedDomain := make(net.IP, len(rawIp))

	copy(resolvedDomain, rawIp)

	return resolvedDomain
}

func PrettyPrint(timestamp time.Time, reply *icmpUtils.IcmpReply, domain string, icmpTime time.Duration) {
	fmt.Printf(
		"[%s] %d bytes from %s (%s): icmp_seq=%d ttl=%d time=%d ms\n",
		timestamp.Format(time.RFC1123Z),
		reply.IpHeader.TotalLength-uint16(len(reply.Reply.Data))+uint16(unsafe.Sizeof(reply.Reply))+uint16(unsafe.Sizeof(reply.IpHeader)),
		domain,
		reply.IpHeader.SrcAddr,
		reply.Reply.SequenceNumber,
		reply.IpHeader.TTL,
		icmpTime.Milliseconds(),
	)
}
