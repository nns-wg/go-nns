package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
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

//go:embed static/*
var staticFiles embed.FS

//go:embed nns-web/build/*
//go:embed nns-web/build/_app/immutable/chunks/*
//go:embed nns-web/build/_app/immutable/entry/*
//go:embed nns-web/build/_app/immutable/assets/*
var svelteFs embed.FS

func InitializeHTTP(errCh chan<- error, dht dhtHost, listenAddr string) {
	http.Handle("/static/", http.FileServer(http.FS(staticFiles)))
	http.HandleFunc("/node-info", infoPage(&dht))
	http.HandleFunc("/", initHandler(&dht))
	log.Printf("Starting HTTP Listener on %s", listenAddr)
	errCh <- http.ListenAndServe(listenAddr, nil)
}

func handleApi(dht *dhtHost, w http.ResponseWriter, r *http.Request) {
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

func handleIndex(svelteBuild *fs.FS, w http.ResponseWriter, r *http.Request) {

	index, err := fs.ReadFile(*svelteBuild, "index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(index)
}

func initHandler(dht *dhtHost) func(w http.ResponseWriter, r *http.Request) {

	svelteBuild, err := fs.Sub(svelteFs, "nns-web/build")
	if err != nil {
		log.Fatal(err)
	}
	svelteFs := http.FS(svelteBuild)
	fileServer := http.FileServer(svelteFs)

	return func(w http.ResponseWriter, r *http.Request) {

		if (r.Header.Get("accept") == "application/json") {
			handleApi(dht, w, r)
			return
		}

		f, err := svelteBuild.Open(r.URL.Path[1:])

		if err != nil {
			handleIndex(&svelteBuild, w, r)
			return
		}

		defer f.Close()
		fileServer.ServeHTTP(w, r)
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

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// Return preflight request for CORS
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

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
		for i := range multiAddresses {
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
