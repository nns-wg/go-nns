package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/multiformats/go-multiaddr"
)

type addrList []multiaddr.Multiaddr
type listenAddrList []string


type Config struct {
	ListenAddr            string
	ProtocolID            string
	DiscoveryPeers        addrList
	Standalone            bool
	DHTPort               string
	PrivateDHTListenAddrs listenAddrList
	PublicDHTListenAddrs  listenAddrList
}

func main() {

	config := parseFlags()

	ctx, cancel := context.WithCancel(context.Background())

	hosts := initializeDHTHosts(ctx, config)

	httpErrCh := make(chan error, 1)
	go InitializeHTTP(httpErrCh, hosts[0], config.ListenAddr)

	run(hosts, httpErrCh, cancel)
}


func parseFlags() Config {
	config := Config{}

	flag.Var(&config.DiscoveryPeers, "peer", "Multiaddress for discovery of NNS peers")
	flag.StringVar(&config.ListenAddr, "http-addr", ":9970", "Listening address for NNS HTTP interface")
	flag.StringVar(&config.DHTPort, "dht-port", "9971", "Listening port for DHT")
	flag.Var(&config.PrivateDHTListenAddrs, "private-addr", "Listening address for NNS DHT interface, in '/ip4/ip-address' or '/ip6/ip-address' format. Specify multiple comma-separated or as multiple parameters.")
	flag.Var(&config.PublicDHTListenAddrs, "public-addr", "Public address for NNS DHT interface, in '/dns/hostname' format. This is used to translate internal listen addresses to a public address that other nodes can connect to. If omitted, the bound private IP addresses will be used.")
	flag.BoolVar(&config.Standalone, "standalone", false, "Run in standalone mode. Will accept saves without connected peers. Use only for debugging.")
	flag.Parse()

	if len(config.PrivateDHTListenAddrs) == 0 {
	  config.PrivateDHTListenAddrs = listenAddrList{"/ip4/0.0.0.0", "/ip6/::"}
	}

	return config
}

func run(hosts []dhtHost, httpErrCh chan error, cancel func()) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-c:
		fmt.Printf("\rExiting...\n")
	case err := <-httpErrCh:
		fmt.Printf("\rError in HTTP Server: %s", err)
	}

	cancel()

	for _, h := range hosts {
		if err := h.host.Close(); err != nil {
			panic(err)
		}
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

func (s *listenAddrList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *listenAddrList) Set(value string) error {
	*s = append(*s, value)
	return nil
}
