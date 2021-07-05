package utils

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/V-H-R-Oliveira/dns-client/protocol"
	"github.com/V-H-R-Oliveira/dns-client/utils"
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
