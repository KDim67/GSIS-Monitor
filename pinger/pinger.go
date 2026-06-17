package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

// PingService pings a service, gathers network tracing metrics and checks SSL + keywords.
func PingService(s Service) *PingLog {
	now := time.Now()
	logEntry := &PingLog{
		Time:            now,
		ServiceID:       s.ID,
		IsHealthy:       false,
		ContentVerified: false,
	}

	// Configure transport with insecure TLS skip option
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: s.SkipTLSVerify,
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		errStr := err.Error()
		logEntry.ErrorMessage = &errStr
		return logEntry
	}

	// Browser Simulation Headers (WAF protection)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "el-GR,el;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Connection", "keep-alive")

	var dnsStart, dnsDone time.Time
	var connStart, connDone time.Time
	var tlsStart, tlsDone time.Time
	var ttfbStart, ttfbDone time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
		},
		ConnectStart: func(_, _ string) {
			connStart = time.Now()
		},
		ConnectDone: func(_, _ string, err error) {
			if err == nil {
				connDone = time.Now()
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err == nil {
				tlsDone = time.Now()
			}
		},
		GotFirstResponseByte: func() {
			ttfbDone = time.Now()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	ttfbStart = time.Now()

	resp, err := client.Do(req)
	totalTime := time.Since(now)

	totalMs := float64(totalTime.Nanoseconds()) / 1e6
	logEntry.TotalResponseMs = &totalMs

	if err != nil {
		errStr := err.Error()
		logEntry.ErrorMessage = &errStr
		return logEntry
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	logEntry.StatusCode = &code

	// Calculate and assign trace values
	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		dnsMs := float64(dnsDone.Sub(dnsStart).Nanoseconds()) / 1e6
		logEntry.DNSLookupMs = &dnsMs
	}
	if !connStart.IsZero() && !connDone.IsZero() {
		connMs := float64(connDone.Sub(connStart).Nanoseconds()) / 1e6
		logEntry.TCPConnectMs = &connMs
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		tlsMs := float64(tlsDone.Sub(tlsStart).Nanoseconds()) / 1e6
		logEntry.TLSHandshakeMs = &tlsMs
	}
	if !ttfbDone.IsZero() {
		ttfbMs := float64(ttfbDone.Sub(ttfbStart).Nanoseconds()) / 1e6
		logEntry.TTFBMs = &ttfbMs
	}

	// Capture SSL certificate expiry remaining days
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		remaining := cert.NotAfter.Sub(time.Now())
		days := int(remaining.Hours() / 24)
		logEntry.SSLExpiryDays = &days
	}

	// Read body for keyword verification
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := fmt.Sprintf("Failed to read body: %s", err.Error())
		logEntry.ErrorMessage = &errStr
		return logEntry
	}
	bodyStr := string(bodyBytes)

	verified := true
	if s.VerificationKeyword != nil && *s.VerificationKeyword != "" {
		kw := *s.VerificationKeyword
		log.Printf("DEBUG [%s]: Checking keyword '%s' (bytes: %v) in body", s.Name, kw, []byte(kw))
		if !strings.Contains(bodyStr, kw) {
			verified = false
			snippet := bodyStr
			if len(bodyStr) > 200 {
				snippet = bodyStr[:200]
			}
			log.Printf("DEBUG [%s]: Keyword not found. Body snippet: %s", s.Name, snippet)
		} else {
			log.Printf("DEBUG [%s]: Keyword found!", s.Name)
		}
	}
	if s.ExclusionKeyword != nil && *s.ExclusionKeyword != "" {
		kw := *s.ExclusionKeyword
		if strings.Contains(bodyStr, kw) {
			verified = false
			log.Printf("DEBUG [%s]: Exclusion keyword '%s' found in body!", s.Name, kw)
		}
	}
	logEntry.ContentVerified = verified

	// Health status evaluation
	if code >= 200 && code < 400 {
		if verified {
			logEntry.IsHealthy = true
		} else {
			logEntry.IsHealthy = false
			errStr := "Content verification failed (keyword verification)"
			logEntry.ErrorMessage = &errStr
		}
	} else {
		logEntry.IsHealthy = false
		errStr := fmt.Sprintf("HTTP Status Code %d", code)
		logEntry.ErrorMessage = &errStr
	}

	return logEntry
}
