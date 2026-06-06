package sync

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	URL       string
	Token     string
	TokenFile string
	CertName  string
	CertFile  string
	KeyFile   string
}

func (c *Config) Validate() error {
	if c.TokenFile != "" {
		data, err := os.ReadFile(c.TokenFile)
		if err != nil {
			return fmt.Errorf("reading token file: %w", err)
		}
		c.Token = strings.TrimSpace(string(data))
	}

	if c.URL == "" {
		return fmt.Errorf("--url or AUTHENTIK_URL is required")
	}
	if c.Token == "" {
		return fmt.Errorf("--token, AUTHENTIK_TOKEN, or --token-file is required")
	}
	if c.CertName == "" {
		return fmt.Errorf("--cert-name is required")
	}
	if c.CertFile == "" {
		return fmt.Errorf("--cert-file is required")
	}
	if c.KeyFile == "" {
		return fmt.Errorf("--key-file is required")
	}
	return nil
}
