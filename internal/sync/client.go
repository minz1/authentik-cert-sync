package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseURL    string
	token      string
	httpClient HTTPClient
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{},
	}
}

type CertKeyPair struct {
	PK   string `json:"pk"`
	Name string `json:"name"`
}

type listResponse struct {
	Count   int           `json:"count"`
	Results []CertKeyPair `json:"results"`
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return req, nil
}

func (c *Client) FindByName(name string) (*CertKeyPair, error) {
	req, err := c.newRequest("GET", "/api/v3/crypto/certificatekeypairs/?name="+name, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET certificatekeypairs returned %d: %s", resp.StatusCode, body)
	}

	var lr listResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("decoding list response: %w", err)
	}

	for _, pair := range lr.Results {
		if pair.Name == name {
			return &pair, nil
		}
	}
	return nil, nil
}

func (c *Client) Create(name, certPEM, keyPEM string) error {
	body, err := buildJSON(name, certPEM, keyPEM)
	if err != nil {
		return err
	}

	req, err := c.newRequest("POST", "/api/v3/crypto/certificatekeypairs/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST certificatekeypairs returned %d: %s", resp.StatusCode, b)
	}
	return nil
}

func (c *Client) Update(pk, name, certPEM, keyPEM string) error {
	body, err := buildJSON(name, certPEM, keyPEM)
	if err != nil {
		return err
	}

	req, err := c.newRequest("PATCH", "/api/v3/crypto/certificatekeypairs/"+pk+"/", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PATCH certificatekeypairs/%s returned %d: %s", pk, resp.StatusCode, b)
	}
	return nil
}

type certKeyPairRequest struct {
	Name            string `json:"name"`
	CertificateData string `json:"certificate_data"`
	KeyData         string `json:"key_data"`
}

func buildJSON(name, certPEM, keyPEM string) (io.Reader, error) {
	data, err := json.Marshal(certKeyPairRequest{Name: name, CertificateData: certPEM, KeyData: keyPEM})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
