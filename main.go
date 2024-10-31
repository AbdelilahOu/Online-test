package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	serverPort = "3430"
)

func main() {
	// parse url becouse NewSingleHostReverseProxy need *url.URL
	wikiUrl, err := url.Parse("https://wikipedia.org")
	if err != nil {
		log.Fatalln("error parsing wikipedia url: ", err)
	}

	// create a proxy
	proxy := httputil.NewSingleHostReverseProxy(wikiUrl)

	// modify headers
	// set host to wikipedia instead of localhost:3430
	oldDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		oldDirector(r)
		r.Host = wikiUrl.Host
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		// if the response isnt an html page do nth
		if !strings.Contains(r.Header.Get("Content-Type"), "text/html") {
			return nil
		}

		// read body
		html, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body.Close()

		// replace wikipedia.org to m-wikipedia.org
		newBody := strings.ReplaceAll(string(html), "wikipedia.org", "m-wikipedia.org")
		log.Println(html)
		log.Println(string(html))
		log.Println(newBody)
		// set new body
		r.Body = io.NopCloser(strings.NewReader(newBody))
		r.Header.Set("Content-Length", fmt.Sprint(len(newBody)))
		r.Header.Set("Content-Type", "text/html")

		return nil
	}

	// handle proxy errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		http.Error(w, "proxy error", http.StatusBadRequest)
	}

	// catch all routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("requesting url: https://wikipedia.org", r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// run server
	fmt.Printf("server is running on port %s\n", serverPort)
	err = http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)
	if err != nil {
		log.Fatalln("server error: ", err)
	}
}
