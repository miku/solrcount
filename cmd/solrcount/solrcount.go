// A proxy for solr requests, that will only reveal the number of results.
//
// Usage of solrcount:
//   -core="biblio": SOLR core name
//   -host="localhost": host of the SOLR server to proxy
//   -listen=":18080": host and port to listen on
//   -port=8080: port of the SOLR server to proxy//

// `host`, `port` and `core` are parameters of the target SOLR server.
// `listen` is a combined `host:port` string, where this proxy should listen.
//
// Starting a server:
//
//     $ solrcount -host 10.0.0.1 -port 8080 -core biblio -listen :9999
//
// This will start a server that listens on `localhost:9999` and will forward
// queries to a SOLR server on `10.0.0.1:8080` using the `biblio` core.
//
// Querying the server:
//
//     $ curl localhost:9999
//     solrcount 1.0.0
//
//     $ curl localhost:9999/proxy?q=Hello
//     {"status":0,"qtime":0,"q":"q=Hello","count":1686}
//
//     $ curl localhost:9999/proxy?q=Hello%20World
//     {"status":0,"qtime":0,"q":"q=Hello","count":1686}
//
//     $ curl localhost:9999/proxy?q=Hello%20OR%20World
//     {"status":0,"qtime":62,"q":"q=Hello%20OR%20World","count":545878}
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/rtt/Go-Solr"
)

const Version = "1.0.0"

type Response struct {
	Status      int    `json:"status"`
	QTime       int    `json:"qtime"`
	QueryString string `json:"q"`
	NumFound    int    `json:"count"`
}

func main() {

	solrHost := flag.String("host", "localhost", "host of the SOLR server to proxy")
	solrPort := flag.Int("port", 8080, "port of the SOLR server to proxy")
	solrCore := flag.String("core", "biblio", "SOLR core name")
	listen := flag.String("listen", ":18080", "host and port to listen on")

	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "solrcount %s\n", Version)
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		s, err := solr.Init(*solrHost, *solrPort, *solrCore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		res, err := s.SelectRaw(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response := Response{Status: res.Status, QTime: res.QTime, QueryString: r.URL.RawQuery, NumFound: res.Results.NumFound}
		b, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "%s", string(b))
	})

	http.ListenAndServe(*listen, nil)
}
