package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Windows Service! Time: %s", time.Now())
	})

	fmt.Println("Starting server on :8075...")
	if err := http.ListenAndServe(":8075", nil); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
