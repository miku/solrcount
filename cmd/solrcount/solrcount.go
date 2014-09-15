// A proxy for solr requests, that will only reveal the number of results.
//
// Usage of solrcount:
//   -core="biblio": SOLR core name
//   -host="localhost": host of the SOLR server to proxy
//   -listen=":18080": host and port to listen on
//   -port=8080: port of the SOLR server to proxy
//   -w=4: concurrency level
//
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
// The query given to the proxy must be already properly escaped. Errors are signalled
// with a HTTP status codes:
//
//     $ curl -v localhost:18080/proxy?q="Hello World"
//     > GET /proxy?q=Hello World HTTP/1.1
//     > User-Agent: curl/7.35.0
//     > Host: 127.0.0.1:18080
//     > Accept: */*
//     >
//     < HTTP/1.1 400 Bad Request
//
// Output can be JSON, XML or TSV, depending on the
// [Accept](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.1) header.
//
//     $ curl -H 'Accept: application/xml' localhost:9999/proxy?q=Hi
//     <response><status>0</status><qtime>1</qtime><q>q=Hi</q><count>4216</count></response>
//
//     $ curl -H 'Accept: text/plain' localhost:9999/proxy?q=Hi
//     4216
//
// Default response mimetype is *application/json*.
//
package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"runtime"

	"github.com/rtt/Go-Solr"
)

const Version = "1.0.2"

const (
	mimeTSV  = "text/tab-separated-values; charset=utf-8"
	mimeXML  = "application/xml; charset=utf-8"
	mimeJSON = "application/json; charset=utf-8"
	mimeText = "text/plain; charset=utf-8"
)

type Response struct {
	XMLName     xml.Name `json:"-"      xml:"response"`
	Status      int      `json:"status" xml:"status"`
	QTime       int      `json:"qtime"  xml:"qtime"`
	QueryString string   `json:"q"      xml:"q"`
	NumFound    int      `json:"count"  xml:"count"`
}

func (r Response) String() string {
	return fmt.Sprintf("%d", r.NumFound)
}

func (r Response) TSV() string {
	return fmt.Sprintf("%d\t%d\t%s\t%d", r.Status, r.QTime, r.QueryString, r.NumFound)
}

func main() {

	solrHost := flag.String("host", "localhost", "host of the SOLR server to proxy")
	solrPort := flag.Int("port", 8080, "port of the SOLR server to proxy")
	solrCore := flag.String("core", "biblio", "SOLR core name")
	listen := flag.String("listen", ":18080", "host and port to listen on")
	numWorkers := flag.Int("w", runtime.NumCPU(), "concurrency level")

	flag.Parse()

	runtime.GOMAXPROCS(*numWorkers)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "solrcount %s, go to: /proxy?q=*:*\n", Version)
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		var err error
		s, err := solr.Init(*solrHost, *solrPort, *solrCore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := s.SelectRaw(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response := Response{
			Status:      res.Status,
			QTime:       res.QTime,
			QueryString: r.URL.RawQuery,
			NumFound:    res.Results.NumFound,
		}
		accept := r.Header.Get("Accept")
		var b []byte
		switch accept {
		case "text/plain":
			b = []byte(fmt.Sprintf("%s\n", response))
			w.Header().Set("Content-Type", mimeText)
		case "text/tab-separated-values":
			b = []byte(fmt.Sprintf("%s\n", response.TSV()))
			w.Header().Set("Content-Type", mimeTSV)
		case "application/xml":
			b, err = xml.Marshal(response)
			w.Header().Set("Content-Type", mimeXML)
		default:
			b, err = json.Marshal(response)
			w.Header().Set("Content-Type", mimeJSON)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s", string(b))
	})
	http.ListenAndServe(*listen, nil)
}
