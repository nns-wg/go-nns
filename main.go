package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	Port           int
	ProtocolID     string
	DiscoveryPeers addrList
}

type addrList []multiaddr.Multiaddr

func main() {

	config := Config{}

	flag.Var(&config.DiscoveryPeers, "peer", "Peer multiaddress for peer discovery")
	flag.IntVar(&config.Port, "port", 0, "")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	host, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Host ID: %s", host.ID().Pretty())
	log.Printf("Connect to me on:")
	for _, addr := range host.Addrs() {
		log.Printf("  %s/p2p/%s", addr, host.ID().Pretty())
	}

	dht, err := NewDHT(ctx, host, config.DiscoveryPeers)
	if err != nil {
		log.Fatal(err)
	}

	//go Discover(ctx, host, dht)

	err = dht.PutValue(ctx, "/unvalidated/hello", []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
	if err != nil {
		log.Printf("put %s", err)
	} else {
		log.Printf("successfully pub?")
	}
	setValue, err := dht.GetValue(ctx, "/unvalidated/hello")
	if err != nil {
		log.Printf("get %s", err)
	}

	log.Printf("The value? %s", setValue)

	run(host, cancel)
}

func run(h host.Host, cancel func()) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c

	fmt.Printf("\rExiting...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := multiaddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}
