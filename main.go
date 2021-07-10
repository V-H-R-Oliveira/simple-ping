package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"github.com/V-H-R-Oliveira/simple-ping/protocol"
	"github.com/V-H-R-Oliveira/simple-ping/utils"
)

func main() {
	userInput := utils.GetUserInput()
	dstIp := net.ParseIP(userInput.Domain)
	var dstIpAddr [4]byte

	if dstIp == nil {
		dstIp = utils.ResolveDomain(userInput.Domain)
		dstIpAddr = utils.IpToByteSlice(dstIp.String())
	} else {
		dstIpAddr = utils.IpToByteSlice(userInput.Domain)
	}

	stats := protocol.NewStatistics()
	icmpRequest := protocol.NewIcmpRequest(dstIp)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	startPingSequence := time.Now()

	go func() {
		defer cancel()

		for i := 1; i <= userInput.Quantity; i++ {
			icmpRequest.Request.SetTimestamp()
			icmpRequest.Request.SetSequenceNumber(uint16(i))
			icmpRequest.Request.ComputeChecksum()

			if i == 1 {
				fmt.Printf(
					"PING %s (%s) %d(%d) bytes of data.\n",
					userInput.Domain,
					dstIp.String(),
					len(icmpRequest.Request.Data),
					len(icmpRequest.Request.Data)+int(unsafe.Sizeof(icmpRequest.IpHeader)),
				)
			}

			requestTime := time.Now()
			stats.IncrementTransmitted()
			icmpReply := icmpRequest.SendRequest(ctx, dstIpAddr)

			if icmpReply != nil {
				stats.IncrementReceived()
				timestamp := protocol.ParseTimestamp(icmpReply.Reply.Data[:32])
				utils.PrettyPrint(timestamp, icmpReply, userInput.Domain, time.Since(requestTime))
			}

			time.Sleep(time.Second)
		}
	}()

	select {
	case <-signalCh:
		stats.SetTotalTime(startPingSequence)
		stats.PrettyStats(userInput.Domain)
	case <-ctx.Done():
		stats.SetTotalTime(startPingSequence)
		stats.PrettyStats(userInput.Domain)
	}
}
