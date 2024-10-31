package main

import (
	"bytes"
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
	// Parse the Wikipedia URL
	wikiUrl, err := url.Parse("https://www.wikipedia.org")
	if err != nil {
		log.Fatalln("error parsing wikipedia url:", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(wikiUrl)

	// Preserve the original director and modify headers
	oldDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		oldDirector(r)
		r.Host = wikiUrl.Host
		// Ensure proper headers for mobile site
		r.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	}

	// Modify response to replace URLs
	proxy.ModifyResponse = func(r *http.Response) error {
		// Check if response is HTML
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(strings.ToLower(contentType), "text/html") {
			return nil
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %v", err)
		}
		r.Body.Close()

		// Replace all variations of Wikipedia URLs
		newBody := strings.ReplaceAll(string(body), "wikipedia", "m-wikipedia")

		// Create new body
		bodyBytes := []byte(newBody)
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		r.ContentLength = int64(len(bodyBytes))
		r.Header.Set("Content-Length", fmt.Sprint(len(bodyBytes)))

		// Ensure proper content type
		if !strings.Contains(contentType, "charset") {
			r.Header.Set("Content-Type", "text/html; charset=utf-8")
		}

		return nil
	}

	// Handle proxy errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
	}

	// Handle all routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxying request: %s%s", wikiUrl.String(), r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	// Start server
	log.Printf("Server running on port %s\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil); err != nil {
		log.Fatalln("server error:", err)
	}
}
