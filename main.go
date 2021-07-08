package main

import (
	"context"
	"fmt"
	"net"
	"time"
	"unsafe"

	"github.com/V-H-R-Oliveira/simple-icmp-client/protocol"
	"github.com/V-H-R-Oliveira/simple-icmp-client/utils"
)

func main() {
	domain := "cloudflare.com" //172.164.0.0 142.250.218.174
	dstIp := net.ParseIP(domain)

	var dstIpAddr [4]byte

	if dstIp == nil {
		dstIp = utils.ResolveDomain(domain)
		dstIpAddr = utils.IpToByteSlice(dstIp.String())
	} else {
		dstIpAddr = utils.IpToByteSlice(domain)
	}

	icmpRequest := protocol.NewIcmpRequest(dstIp)
	ctx, cancel := context.WithTimeout(context.Background(), protocol.DEFAULT_TIMEOUT*time.Second)
	defer cancel()

	for i := 1; i <= 5; i++ {

		icmpRequest.Request.SetTimestamp()
		icmpRequest.Request.SetSequenceNumber(uint16(i))
		icmpRequest.Request.ComputeChecksum()

		if i == 1 {
			fmt.Printf(
				"PING %s (%s) %d(%d) bytes of data.\n",
				domain,
				dstIp.String(),
				len(icmpRequest.Request.Data),
				len(icmpRequest.Request.Data)+int(unsafe.Sizeof(icmpRequest.IpHeader)),
			)
		}

		requestTime := time.Now()
		icmpRequest.SendRequest(dstIpAddr)
		icmpReply := protocol.GetReply(ctx, len(icmpRequest.Request.Data)+int(unsafe.Sizeof(icmpRequest.IpHeader)))

		if icmpReply != nil {
			timestamp := protocol.ParseTimestamp(icmpReply.Reply.Data[:32])
			utils.PrettyPrint(timestamp, icmpReply, domain, time.Since(requestTime))
		}

		time.Sleep(time.Second)
	}
}
