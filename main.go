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

type Config struct {
	ListenAddr     string
	ProtocolID     string
	DiscoveryPeers addrList
	Standalone     bool
}

func main() {

	config := parseFlags()

	ctx, cancel := context.WithCancel(context.Background())

	hosts := initializeDHTHosts(ctx, config)

	httpErrCh := make(chan error, 1)
	go InitializeHTTP(httpErrCh, hosts[0].dht, config.ListenAddr)

	run(hosts, httpErrCh, cancel)
}

func parseFlags() Config {
	config := Config{}

	flag.Var(&config.DiscoveryPeers, "peer", "Multiaddress for discovery of NNS peers")
	flag.StringVar(&config.ListenAddr, "addr", ":3333", "Listening address for NNS HTTP interface")
	flag.BoolVar(&config.Standalone, "standalone", false, "Run in standalone mode. Will accept saves without peers.")
	flag.Parse()

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
