package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/govice/golinks/blockchain"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	net "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"
)

// var ctx context.Context
//todo see https://gist.github.com/upperwal/38cd0c98e4a6b34c061db0ff26def9b9#file-libp2p_chat_bootstrapping-md
//todo see https://github.com/ailabstw/go-pttai/issues/97
func handleStream(stream net.Stream) {
	log.Println("Got a new stream")
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readData(rw)
	go writeData(rw)
}

var ErrPeerPort = errors.New("peer port not assigned")

func startPeer(ctx context.Context) error {
	if !viper.IsSet("peer_port") {
		return ErrPeerPort
	}
	hostAddress, err := maddr.NewMultiaddr("/ip4/0.0.0.0/tcp/" + viper.GetString("peer_port"))
	if err != nil {
		return err
	}

	var config = Config{
		ListenAddresses:  []maddr.Multiaddr{hostAddress},
		ProtocolID:       "/golinks/0.0.1",
		BootstrapPeers:   dht.DefaultBootstrapPeers,
		RendezvousString: "golinks-daemon",
	}

	prvKey, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)

	if err != nil {
		return err
	}

	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(hostAddress),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		return err
	}

	log.Println("Host created. We are:", host.ID())
	log.Println(host.Addrs())

	// Set a function as stream handler.
	host.SetStreamHandler(protocol.ID(config.ProtocolID), handleStream)

	// go runDHT(ctx, host, config)
	if err := runMDNS(ctx, host, config); err != nil {
		return err
	}

	return nil
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			time.Sleep(time.Second * 10)
			continue
		}

		if str == "" || str == "\n" {
			time.Sleep(time.Second * 10)
			continue
		}

		if str != "\n" {
			var peerChain blockchain.Blockchain
			err = json.Unmarshal([]byte(str), &peerChain)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 10)
				continue
			}

			gci, err := blockchainService.GCI(&peerChain)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 10)
				continue
			}
			log.Println("GCI: ", gci)

			if peerChain.Length() > blockchainService.ChainLength() && gci >= 0 {
				if err := blockchainService.UpdateChain(&peerChain); err != nil {
					log.Println(err)
					time.Sleep(time.Second * 10)
					continue
				}
				log.Println("UPDATED CHAIN")
			}
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	for {
		bytes, err := blockchainService.ChainJSON()
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 10)
			continue
		}

		_, err = rw.WriteString(string(bytes) + "\n")
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 10)
			continue
		}

		err = rw.Flush()
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 10)
			continue
		}

		time.Sleep(time.Second * 10)
	}
}
