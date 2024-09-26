package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/webdav"
)

func main() {
	handler := &webdav.Handler{
		FileSystem: webdav.Dir("./webdav_root"),
		LockSystem: webdav.NewMemLS(),
	}

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})

	mux.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Quit request received. Shutting down server...")
		w.Write([]byte("Server is shutting down..."))
		go func() {
			// Wait a moment to allow the response to be sent
			time.Sleep(100 * time.Millisecond)
			// Shutdown the server
			if err := server.Shutdown(context.Background()); err != nil {
				log.Printf("Error during server shutdown: %v", err)
			}
		}()
	})

	if err := os.MkdirAll("./webdav_root", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting WebDAV server on http://localhost:8080")
	fmt.Println("Access http://localhost:8080/quit to stop the server")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}

	fmt.Println("Server has been shut down")
}

//
