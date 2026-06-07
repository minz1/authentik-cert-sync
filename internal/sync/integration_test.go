package sync

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestIntegrationCreateThenUpdate(t *testing.T) {
	const certName = "ldap-outpost"
	const fakePK = "uuid-0001"

	var stored *CertKeyPair

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/crypto/certificatekeypairs/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			results := []CertKeyPair{}
			if stored != nil {
				results = append(results, *stored)
			}
			json.NewEncoder(w).Encode(map[string]any{"count": len(results), "results": results})

		case http.MethodPost:
			var req certKeyPairRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			stored = &CertKeyPair{PK: fakePK, Name: req.Name}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(stored)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v3/crypto/certificatekeypairs/"+fakePK+"/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req certKeyPairRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stored.Name = req.Name
		json.NewEncoder(w).Encode(stored)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	certFile := dir + "/cert.pem"
	keyFile := dir + "/key.pem"
	if err := os.WriteFile(certFile, []byte("CERT-DATA"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, []byte("KEY-DATA"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		URL:      srv.URL,
		Token:    "test-token",
		CertName: certName,
		CertFile: certFile,
		KeyFile:  keyFile,
	}

	// First run: cert doesn't exist → POST (create)
	client := NewClient(cfg.URL, cfg.Token)
	if err := Run(client, cfg); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if stored == nil || stored.PK != fakePK {
		t.Fatal("expected cert to be created")
	}

	// Second run: cert exists → PATCH (update)
	if err := Run(client, cfg); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
}
