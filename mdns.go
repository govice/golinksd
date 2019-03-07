package main

import (
	"bufio"
	"context"
	"fmt"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

type discoveryNotifee struct {
	PeerChan chan pstore.PeerInfo
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi pstore.PeerInfo) {
	n.PeerChan <- pi
}

func runMDNS(ctx context.Context, host host.Host, config Config) {
	ser, err := discovery.NewMdnsService(ctx, host, time.Hour, config.RendezvousString)
	if err != nil {
		panic(err)
	}

	n := &discoveryNotifee{}
	n.PeerChan = make(chan pstore.PeerInfo)

	ser.RegisterNotifee(n)

	for peer := range n.PeerChan {

		fmt.Println("Found peer:", peer, ", connecting")

		if err := host.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
		}

		// open a stream, this stream will be handled by handleStream other end
		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

		if err != nil {
			fmt.Println("Stream open failed", err)
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeData(rw)
			go readData(rw)
			fmt.Println("Connected to:", peer)
		}
	}
	select {} //wait here
}
