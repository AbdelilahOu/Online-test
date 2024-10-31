package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	serverPort = "3430"
)

func main() {
	fmt.Println("setup  done!!")

	// catch all routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)

		w.Write([]byte("Hellooo"))
	})

	fmt.Printf("server is running on port %s\n", serverPort)

	err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)
	if err != nil {
		log.Fatalln("server error: ", err)
	}
}
