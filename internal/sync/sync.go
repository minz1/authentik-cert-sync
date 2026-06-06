package sync

import (
	"fmt"
	"os"
)

type SyncClient interface {
	FindByName(name string) (*CertKeyPair, error)
	Create(name, certPEM, keyPEM string) error
	Update(pk, name, certPEM, keyPEM string) error
}

func Run(client SyncClient, cfg Config) error {
	certPEM, err := os.ReadFile(cfg.CertFile)
	if err != nil {
		return fmt.Errorf("reading cert file: %w", err)
	}

	keyPEM, err := os.ReadFile(cfg.KeyFile)
	if err != nil {
		return fmt.Errorf("reading key file: %w", err)
	}

	existing, err := client.FindByName(cfg.CertName)
	if err != nil {
		return fmt.Errorf("looking up certificate: %w", err)
	}

	if existing == nil {
		if err := client.Create(cfg.CertName, string(certPEM), string(keyPEM)); err != nil {
			return fmt.Errorf("creating certificate: %w", err)
		}
		fmt.Printf("created certificate %q\n", cfg.CertName)
	} else {
		if err := client.Update(existing.PK, cfg.CertName, string(certPEM), string(keyPEM)); err != nil {
			return fmt.Errorf("updating certificate: %w", err)
		}
		fmt.Printf("updated certificate %q (pk=%s)\n", cfg.CertName, existing.PK)
	}

	return nil
}
