package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/minz1/authentik-cert-sync/internal/sync"
)

func main() {
	var cfg sync.Config

	flag.StringVar(&cfg.URL, "url", os.Getenv("AUTHENTIK_URL"), "Authentik base URL")
	flag.StringVar(&cfg.Token, "token", os.Getenv("AUTHENTIK_TOKEN"), "Authentik API token")
	flag.StringVar(&cfg.TokenFile, "token-file", "", "Path to file containing Authentik API token")
	flag.StringVar(&cfg.CertName, "cert-name", "", "Certificate name in Authentik")
	flag.StringVar(&cfg.CertFile, "cert-file", "", "Path to certificate PEM file")
	flag.StringVar(&cfg.KeyFile, "key-file", "", "Path to private key PEM file")
	flag.Parse()

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	client := sync.NewClient(cfg.URL, cfg.Token)
	if err := sync.Run(client, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
