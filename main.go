package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
)

const (
	serverPort = "3430"
)

var (
	redirectStatusCodes = []int{301, 302, 307, 308}
)

func main() {
	// Parse the Wikipedia URL
	wikiUrl, err := url.Parse("https://wikipedia.org")
	if err != nil {
		log.Fatalln("error parsing wikipedia url:", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(wikiUrl)

	// Preserve the original director and modify headers
	oldDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		oldDirector(r)
		r.URL.Scheme = wikiUrl.Scheme
		r.Host = wikiUrl.Host
		// Add headers to ensure proper encoding
		r.Header.Set("Accept-Encoding", "identity")
		r.Header.Set("Accept-Charset", "utf-8")
	}

	// Modify response to replace URLs and handle redirects
	proxy.ModifyResponse = func(r *http.Response) error {
		// Handle redirects
		if slices.Contains(redirectStatusCodes, r.StatusCode) {
			location := r.Header.Get("Location")
			if location != "" {
				body, err := fetchRedirectLocation(location)
				if err != nil {
					return err
				}
				newBody := processHtml(string(body))
				r.StatusCode = 200
				r.Status = "200 OK"
				r.Header.Del("Location")
				r.Body = io.NopCloser(bytes.NewReader([]byte(newBody)))
				r.ContentLength = int64(len(newBody))
				r.Header.Set("Content-Length", fmt.Sprint(len(newBody)))

				// Ensure proper content type with UTF-8 charset
				contentType := r.Header.Get("Content-Type")
				if !strings.Contains(contentType, "charset") {
					r.Header.Set("Content-Type", "text/html; charset=utf-8")
				}

				return nil
			}
		}

		// Check if response is HTML
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(strings.ToLower(contentType), "text/html") {
			// Read body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				return fmt.Errorf("error reading response body: %v", err)
			}
			r.Body.Close()

			bodyStr := processHtml(string(body))

			// Create new body
			bodyBytes := []byte(bodyStr)
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			r.ContentLength = int64(len(bodyBytes))
			r.Header.Set("Content-Length", fmt.Sprint(len(bodyBytes)))
			if !strings.Contains(contentType, "charset") {
				r.Header.Set("Content-Type", "text/html; charset=utf-8")
			}
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

func fetchRedirectLocation(url string) ([]byte, error) {
	log.Println("Redirected to :", url)
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching content: %v", err)
	}
	defer resp.Body.Close()
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return body, nil
}

func processHtml(html string) string {
	// Replace all variations of Wikipedia URLs
	replacements := map[string]string{
		".wikipedia.org": ".m-wikipedia.org",
	}

	newBody := html
	for old, new := range replacements {
		newBody = strings.ReplaceAll(newBody, old, new)
	}

	newBody = strings.Replace(
		newBody,
		"</body>",
		`
			<script>
				document.addEventListener('DOMContentLoaded', function () {
					// link effect
					const links = document.querySelectorAll('a');
					links.forEach((link) => {
						link.style.transition = 'all 0.3s ease';
						link.addEventListener('mouseenter', () => {
							link.style.backgroundColor = '#ffeb3b';
							link.style.textDecoration = 'none';
							link.style.borderRadius = '3px';
							link.style.border = '1px solid orange';
							link.style.padding = '0 4px';
						});
						link.addEventListener('mouseleave', () => {
							link.style.backgroundColor = 'transparent';
							link.style.padding = '0';
							link.style.border = '';
						});
					});
				});
			</script>
			</body>
		`,
		1,
	)
	return newBody
}
