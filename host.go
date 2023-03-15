package main

import (
	"context"
	"fmt"
	"log"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

type dhtHost struct {
	host host.Host
	dht *dht.IpfsDHT
}

func initializeDHTHosts(ctx context.Context, config Config) []dhtHost {
	var hosts []dhtHost

	host, dht := initializeDHT(ctx, config.DiscoveryPeers)
	hosts = append(hosts, dhtHost{host, dht})

	if config.Standalone {
		/*
		 * Initialize a SECOND dht node in this process because libp2p only works if two nodes are present.
		 * There's probably some way to make the HTTP server work on the first node, but I'm not quite sure
		 * what that is yet. FIXME this shouldn't be necessary (but probably requires
		 * modification of libp2p to fix).
		 */
		hostAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", host.Addrs()[0], host.ID().Pretty()))
		if err != nil {
			log.Fatal(err)
		}
		hostAddrs := []multiaddr.Multiaddr{hostAddr}
		host2, dht2 := initializeDHT(ctx, hostAddrs)
		hosts = append(hosts, dhtHost{host2, dht2})
	}

  return hosts
}