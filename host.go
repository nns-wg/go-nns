package main

import (
	"context"
	"fmt"
	"log"

	"github.com/multiformats/go-multiaddr"
)

func initializeDHTHosts(ctx context.Context, config Config) []dhtHost {
	var hosts []dhtHost

	nnsConfig := NNSConfig{
		Port:               config.DHTPort,
		PrivateListenAddrs: config.PrivateDHTListenAddrs,
		PublicListenAddrs:  config.PublicDHTListenAddrs,
	}
	host := initializeDHT(ctx, nnsConfig, config.DiscoveryPeers)
	hosts = append(hosts, host)

	if config.Standalone {
		/*
		 * Initialize a SECOND dht node in this process because libp2p only works if two nodes are present.
		 * There's probably some way to make the HTTP server work on the first node, but I'm not quite sure
		 * what that is yet. FIXME this shouldn't be necessary (but probably requires
		 * modification of libp2p to fix).
		 */
		hostAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", host.host.Addrs()[0], host.host.ID().Pretty()))
		if err != nil {
			log.Fatal(err)
		}

		nnsConfig := NNSConfig{
			Port:               "0",
			PrivateListenAddrs: []string{"/ip4/127.0.0.1", "/ip6/::1"},
			PublicListenAddrs:  []string{"/ip4/127.0.0.1", "ip6/::1"},
		}
		hostAddrs := []multiaddr.Multiaddr{hostAddr}
		host = initializeDHT(ctx, nnsConfig, hostAddrs)
		hosts = append(hosts, host)
	}

	return hosts
}
