package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Helper to create string pointers
func strPtr(s string) *string {
	return &s
}

func TestPingService_Success(t *testing.T) {
	// Start a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Welcome to GSIS Dashboard Portal</body></html>"))
	}))
	defer server.Close()

	service := Service{
		ID:                  1,
		Name:                "Test Service Success",
		URL:                 server.URL,
		VerificationKeyword: strPtr("Dashboard"),
		ExclusionKeyword:    strPtr("error"),
		SkipTLSVerify:       false,
	}

	log := PingService(service)

	if !log.IsHealthy {
		t.Errorf("Expected service to be healthy, got unhealthy. Error: %v", getStrVal(log.ErrorMessage))
	}
	if log.StatusCode == nil || *log.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %v", http.StatusOK, log.StatusCode)
	}
	if !log.ContentVerified {
		t.Error("Expected content to be verified")
	}
	if log.TotalResponseMs == nil || *log.TotalResponseMs <= 0 {
		t.Errorf("Expected total response time > 0, got %v", log.TotalResponseMs)
	}
	if log.TTFBMs == nil || *log.TTFBMs <= 0 {
		t.Errorf("Expected TTFB time > 0, got %v", log.TTFBMs)
	}
}

func TestPingService_KeywordMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Internal Portal Page</body></html>"))
	}))
	defer server.Close()

	service := Service{
		ID:                  2,
		Name:                "Test Service Keyword Missing",
		URL:                 server.URL,
		VerificationKeyword: strPtr("GSIS"),
		SkipTLSVerify:       false,
	}

	log := PingService(service)

	if log.IsHealthy {
		t.Error("Expected service to be unhealthy due to missing keyword")
	}
	if log.ContentVerified {
		t.Error("Expected content verification to fail")
	}
	if log.ErrorMessage == nil || !strings.Contains(*log.ErrorMessage, "keyword verification") {
		t.Errorf("Expected error message relating to keyword verification, got %v", log.ErrorMessage)
	}
}

func TestPingService_ExclusionKeywordFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Database maintenance error occurred</body></html>"))
	}))
	defer server.Close()

	service := Service{
		ID:                  3,
		Name:                "Test Service Exclusion Found",
		URL:                 server.URL,
		ExclusionKeyword:    strPtr("error"),
		SkipTLSVerify:       false,
	}

	log := PingService(service)

	if log.IsHealthy {
		t.Error("Expected service to be unhealthy due to exclusion keyword match")
	}
	if log.ContentVerified {
		t.Error("Expected content verification to fail")
	}
}

func TestPingService_HTTP404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Page not found"))
	}))
	defer server.Close()

	service := Service{
		ID:            4,
		Name:          "Test Service 404",
		URL:           server.URL,
		SkipTLSVerify: false,
	}

	log := PingService(service)

	if log.IsHealthy {
		t.Error("Expected service to be unhealthy due to 404 status code")
	}
	if log.StatusCode == nil || *log.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %v", log.StatusCode)
	}
	if log.ErrorMessage == nil || !strings.Contains(*log.ErrorMessage, "HTTP Status Code 404") {
		t.Errorf("Expected error message relating to HTTP status code, got %v", log.ErrorMessage)
	}
}

func TestPingService_TLSAndSSL(t *testing.T) {
	// Start a mock TLS (HTTPS) server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Secure Area"))
	}))
	defer server.Close()

	// In test environment, the server's certificate is signed by a mock CA.
	// We must configure TLS client trace to skip verification or use the mock client.
	service := Service{
		ID:            5,
		Name:          "Test TLS Service",
		URL:           server.URL,
		SkipTLSVerify: true, // Bypass certificate verification since it's a test CA
	}

	// Make sure we trust the mock server TLS config if needed, or SkipTLSVerify will handle it
	log := PingService(service)

	if !log.IsHealthy {
		t.Errorf("Expected TLS service to be healthy, got unhealthy. Error: %v", getStrVal(log.ErrorMessage))
	}
	if log.TLSHandshakeMs == nil || *log.TLSHandshakeMs <= 0 {
		// TLS handshake should be recorded for HTTPS
		t.Errorf("Expected TLS handshake time > 0, got %v", log.TLSHandshakeMs)
	}
	if log.SSLExpiryDays == nil {
		t.Error("Expected SSL certificate expiry days to be recorded")
	} else if *log.SSLExpiryDays < 0 {
		t.Logf("Warning: SSL Expiry days is negative: %d", *log.SSLExpiryDays)
	}
}
