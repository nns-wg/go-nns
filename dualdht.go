package main

import (
	"context"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/multiformats/go-multiaddr"
)

func NewDualDHT(ctx context.Context, host host.Host, bootstrapPeers []multiaddr.Multiaddr) (*dual.DHT, error) {
	var options []dual.Option

	/*
	log.Printf("%v", bootstrapPeers)
	if len(bootstrapPeers) == 0 {
		options = append(options, dual.DHTOption(dht.Mode(dht.ModeServer)))
	}
	*/

	options = append(options, dual.DHTOption(dht.NamespacedValidator("unvalidated", Validator{})))
	options = append(options, dual.DHTOption(dht.ProtocolPrefix("/nns")))

	kdht, err := dual.New(ctx, host, options...)
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