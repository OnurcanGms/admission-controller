package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

var (
	tlsCertFile string = ""
	tlsKeyFile  string = ""
)

func main() {

	// Load the tls files from env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	tlsCertFile := os.Getenv("TLS_CERT_FILE")
	tlsKeyFile := os.Getenv("TLS_KEY_FILE")
	if tlsCertFile == "" || tlsKeyFile == "" {
		log.Fatalf("TLS_CERT_FILE and TLS_KEY_FILE must be set in the .env file")
	}
	// mux for handling request
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", validePod)
	// creating a new server which is listening port 443
	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
	}
	// starting new server on port 443 with tls files
	log.Println(server.ListenAndServeTLS(tlsCertFile, tlsKeyFile))
}
