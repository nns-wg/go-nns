package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func NewDHT(ctx context.Context, host host.Host, bootstrapPeers []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {
	var options []dht.Option

	options = append(
		options,
		dht.Mode(dht.ModeAutoServer),
		dht.Validator(Validator{ctx: ctx}),
		dht.ProtocolPrefix("/nns"))

	kdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}

	if err = kdht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, peerAddr := range bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				log.Printf("Error while connecting to node %q: %-v", peerinfo, err)
			} else {
				log.Printf("Connection established with bootstrap node: %q", *peerinfo)
			}
		}()
	}
	wg.Wait()

	return kdht, nil
}

type NNSConfig struct {
	Port               string
	PrivateListenAddrs []string
	PublicListenAddrs  []string
}

// This is lifted from the libp2p defaults.
var listenAddrTemplates = []string{
	"%s/tcp/%s",
	"%s/udp/%s/quic",
	"%s/udp/%s/quic-v1",
	"%s/udp/%s/quic-v1/webtransport",
}

type dhtHost struct {
	host host.Host
	dht  *dht.IpfsDHT
	pubAddrs []string
}

func initializeDHT(ctx context.Context, config NNSConfig, discoveryPeers addrList) (dhtHost) {

	var privAddrs []string

	for _, listenAddrTemplate := range listenAddrTemplates {
		for _, listenAddr := range config.PrivateListenAddrs {
			privAddrs = append(privAddrs, fmt.Sprintf(listenAddrTemplate, listenAddr, config.Port))
		}
	}

	host, err := libp2p.New(libp2p.ListenAddrStrings(privAddrs...))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Host ID: %s", host.ID().Pretty())
	log.Printf("Connect to me on:")
	for _, addr := range host.Addrs() {
		log.Printf("  %s/p2p/%s", addr, host.ID().Pretty())
	}

	dht, err := NewDHT(ctx, host, discoveryPeers)
	if err != nil {
		log.Fatal(err)
	}

	return dhtHost{host, dht, config.PublicListenAddrs}
}
