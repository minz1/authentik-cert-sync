package sync

import (
	"errors"
	"os"
	"testing"
)

type mockClient struct {
	findResult *CertKeyPair
	findErr    error
	createErr  error
	updateErr  error

	createCalled bool
	updateCalled bool
	updatePK     string
}

func (m *mockClient) FindByName(name string) (*CertKeyPair, error) {
	return m.findResult, m.findErr
}

func (m *mockClient) Create(name, certPEM, keyPEM string) error {
	m.createCalled = true
	return m.createErr
}

func (m *mockClient) Update(pk, name, certPEM, keyPEM string) error {
	m.updateCalled = true
	m.updatePK = pk
	return m.updateErr
}

func makeTempFiles(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	cert := dir + "/cert.pem"
	key := dir + "/key.pem"
	if err := os.WriteFile(cert, []byte("CERT"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(key, []byte("KEY"), 0600); err != nil {
		t.Fatal(err)
	}
	return Config{CertName: "test", CertFile: cert, KeyFile: key}
}

func TestRunCreatesWhenNotFound(t *testing.T) {
	cfg := makeTempFiles(t)
	m := &mockClient{findResult: nil}
	if err := Run(m, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.createCalled {
		t.Error("expected Create to be called")
	}
	if m.updateCalled {
		t.Error("expected Update not to be called")
	}
}

func TestRunUpdatesWhenFound(t *testing.T) {
	cfg := makeTempFiles(t)
	m := &mockClient{findResult: &CertKeyPair{PK: "abc-123", Name: "test"}}
	if err := Run(m, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.createCalled {
		t.Error("expected Create not to be called")
	}
	if !m.updateCalled {
		t.Error("expected Update to be called")
	}
	if m.updatePK != "abc-123" {
		t.Errorf("expected pk=abc-123, got %q", m.updatePK)
	}
}

func TestRunPropagatesFindError(t *testing.T) {
	cfg := makeTempFiles(t)
	m := &mockClient{findErr: errors.New("network error")}
	if err := Run(m, cfg); err == nil {
		t.Error("expected error")
	}
}
