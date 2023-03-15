package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func InitializeHTTP(errCh chan<- error, dht *dht.IpfsDHT, listenAddr string) {
	http.HandleFunc("/", initHandler(dht))
	log.Printf("Starting HTTP Listener on %s", listenAddr)
	errCh <- http.ListenAndServe(listenAddr, nil)
}

func initHandler(dht *dht.IpfsDHT) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handleSet(dht, w, r)
		case "GET":
			handleGet(dht, w, r)
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

	dhtPath := name
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

	name := r.URL.Path
	val, err := dht.GetValue(r.Context(), name)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, fmt.Sprintf("Couldn't find %s", name))
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(val[:]))
}
