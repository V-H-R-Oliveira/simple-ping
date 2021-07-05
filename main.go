package main

import (
	"fmt"
	"net"
	"time"
	"unsafe"

	"github.com/V-H-R-Oliveira/simple-icmp-client/protocol"
	"github.com/V-H-R-Oliveira/simple-icmp-client/utils"
)

func prettyPrint(reply *protocol.IcmpReply, domain string, icmpTime time.Duration) {
	fmt.Printf(
		"%d bytes from %s (%s): icmp_seq=%d ttl=%d time=%d ms\n",
		len(reply.Reply.Data), // fix that
		domain,
		reply.IpHeader.SrcAddr,
		reply.Reply.SequenceNumber,
		reply.IpHeader.TTL,
		icmpTime.Milliseconds(), // fix that
	)
}

func main() {
	domain := "142.250.218.174" //172.164.0.0
	dstIp := net.ParseIP(domain)

	var dstIpAddr [4]byte

	if dstIp == nil {
		dstIp = utils.ResolveDomain(domain)
		dstIpAddr = utils.IpToByteSlice(dstIp.String())
	} else {
		dstIpAddr = utils.IpToByteSlice(domain)
	}

	icmpRequest := protocol.NewIcmpRequest(dstIp)

	fmt.Println(len(icmpRequest.Request.Data), int(unsafe.Sizeof(icmpRequest.IpHeader)))

	fmt.Printf(
		"PING %s (%s) %d(%d) bytes of data.\n",
		domain,
		dstIp.String(),
		len(icmpRequest.Request.Data),
		len(icmpRequest.Request.Data)+int(unsafe.Sizeof(icmpRequest.IpHeader)),
	)

	for i := 1; i <= 5; i++ {
		icmpRequest.Request.SetSequenceNumber(uint16(i))
		startPing := time.Now()

		icmpRequest.SendRequest(dstIpAddr)
		reply := protocol.GetReply()

		if reply != nil {
			prettyPrint(reply, domain, time.Since(startPing))
		}

		icmpRequest.Request.ClearChecksum()
		time.Sleep(time.Second)
	}

}
