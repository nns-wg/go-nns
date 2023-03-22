package main

import (
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"text/template"

	dht "github.com/libp2p/go-libp2p-kad-dht"
)

//go:embed templates/status.html
var statusTemplate embed.FS

//go:embed templates/index.html
var indexTemplate embed.FS

//go:embed static/*
var staticFiles embed.FS

func InitializeHTTP(errCh chan<- error, dht dhtHost, listenAddr string) {
	http.Handle("/static/", http.FileServer(http.FS(staticFiles)))
	http.HandleFunc("/node-info", infoPage(&dht))
	http.HandleFunc("/", initHandler(dht))
	log.Printf("Starting HTTP Listener on %s", listenAddr)
	errCh <- http.ListenAndServe(listenAddr, nil)
}

func initHandler(dht dhtHost) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// Serve the index page.
			indexPage(w, r)
			return
		}

		switch r.Method {
		case "POST":
			handleSet(dht.dht, w, r)
		case "GET":
			handleGet(dht.dht, w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Invalid request.")
		}
	}
}

func handleSet(dht *dht.IpfsDHT, w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %s", err)
	}

	dhtPath := name[1:]
	log.Printf("Storing %s", dhtPath)
	err = dht.PutValue(r.Context(), dhtPath, body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf("Failed to save: %s", err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	io.WriteString(w, "OK")
}

func handleGet(dht *dht.IpfsDHT, w http.ResponseWriter, r *http.Request) {

	// strip leading '/'
	name := r.URL.Path[1:]
	val, err := dht.GetValue(r.Context(), name)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, fmt.Sprintf("Couldn't find %s", name))
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(val[:]))
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(indexTemplate, "templates/index.html")
	if err != nil {
		log.Print("Error parsing template: ", err)
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Print("Error displaying index page: ", err)
	}
}

type StatusPageData struct {
	HostId         string
	MultiAddresses []string
}

func infoPage(dhtHost *dhtHost) func(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(statusTemplate, "templates/status.html")
	if err != nil {
		log.Print("Error parsing template: ", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {

		var multiAddresses []string
		if len(dhtHost.pubAddrs) > 0 {

			// We need to substitute the public addresses for the internal ones.
			for _, mAddr := range dhtHost.host.Addrs() {
				parts := strings.Split(mAddr.String(), "/")
				pubAddr := strings.Join(parts[3:], "/")
				for _, addr := range dhtHost.pubAddrs {
					multiAddresses = append(multiAddresses,
						addr+"/"+pubAddr+"/p2p/"+dhtHost.host.ID().String(),
					)
				}
			}
		} else {
			for _, mAddr := range dhtHost.host.Addrs() {
				multiAddresses = append(multiAddresses,
					mAddr.String()+"/p2p/"+dhtHost.host.ID().String(),
				)
			}
		}

		sort.Strings(multiAddresses)
		result := make([]string, 0, len(multiAddresses))
		for i, _ := range multiAddresses {
			if i == 0 || multiAddresses[i] != multiAddresses[i-1] {
				result = append(result, multiAddresses[i])
			}
		}

		multiAddresses = result

		data := StatusPageData{
			HostId:         dhtHost.host.ID().String(),
			MultiAddresses: multiAddresses,
		}

		err = t.Execute(w, data)
		if err != nil {
			log.Print("Error displaying status page: ", err)
		}
	}
}
