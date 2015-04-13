package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var toURL *url.URL
var proxy *httputil.ReverseProxy

// func noOpServiceWorker(w http.ResponseWriter) {
// 	w.Header().Add("Content-Type", "text/javascript")
// 	w.Write([]byte("self.onload = function() {};\n"))
// }

func handler(w http.ResponseWriter, r *http.Request) {
	//if (r.URL.Path == "/worker.js") {
	//	noOpServiceWorker(w)
	//	return
	//}

	r.Host = ""
	proxy.ServeHTTP(w, r)
}

func main() {
	var port int
	var to string
	flag.StringVar(&to, "to", "", "The URL to direct requests to")
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.Parse()

	var err error
	toURL, err = url.Parse(to)
	if err != nil {
		log.Fatal(err)
	}
	proxy = httputil.NewSingleHostReverseProxy(toURL)
	transport := http.DefaultTransport.(*http.Transport)
	proxy.Transport = transport
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	http.HandleFunc("/", handler)
	log.Print("Starting...")
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
