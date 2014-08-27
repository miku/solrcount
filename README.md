solrcount
=========

A proxy for solr requests, that will only reveal the number of results.

    $ solrcount -h
    Usage of solrcount:
      -core="biblio": SOLR core name
      -host="localhost": host of the SOLR server to proxy
      -listen=":18080": host and port to listen on
      -port=8080: port of the SOLR server to proxy

`host`, `port` and `core` are parameters of the target SOLR server.
`listen` is a combined `host:port` string, where this proxy should listen.

Starting a server:

    $ solrcount -host 10.0.0.1 -port 8080 -core biblio -listen :9999

This will start a server that listens on `localhost:9999` and will forward
queries to a SOLR server on `10.0.0.1:8080` using the `biblio` core.

Querying the server:

    $ curl localhost:9999
    solrcount 1.0.0

    $ curl localhost:9999/proxy?q=Hello
    {"status":0,"qtime":0,"q":"q=Hello","count":1686}

    $ curl localhost:9999/proxy?q=Hello%20World
    {"status":0,"qtime":0,"q":"q=Hello","count":1686}

    $ curl localhost:9999/proxy?q=Hello%20OR%20World
    {"status":0,"qtime":62,"q":"q=Hello%20OR%20World","count":545878}

The query given to the proxy must be already properly escaped. Errors are signalled
with a HTTP status codes:

    $ curl -v localhost:18080/proxy?q="Hello World"
    > GET /proxy?q=Hello World HTTP/1.1
    > User-Agent: curl/7.35.0
    > Host: 127.0.0.1:18080
    > Accept: */*
    >
    < HTTP/1.1 400 Bad Request

Output can be JSON, XML or TSV, depending on the [Accept](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.1) header.

    $ curl -H 'accept: application/xml' localhost:9999/proxy?q=Hello
    <response><status>0</status><qtime>0</qtime><q>q=Hello</q><count>1686</count></response>

    $ curl -H 'accept: text/plain' localhost:9999/proxy?q=Hello
	1686

Default response mimetype is *application/json*.
