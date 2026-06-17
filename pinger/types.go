package main

import (
	"time"
)

// Service represents a monitored service.
type Service struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	URL                 string    `json:"url"`
	VerificationKeyword *string   `json:"verification_keyword"`
	ExclusionKeyword    *string   `json:"exclusion_keyword"`
	SkipTLSVerify       bool      `json:"skip_tls_verify"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// PingLog represents telemetry recorded after a ping.
type PingLog struct {
	Time            time.Time `json:"time"`
	ServiceID       int       `json:"service_id"`
	IsHealthy       bool      `json:"is_healthy"`
	StatusCode      *int      `json:"status_code"`
	ErrorMessage    *string   `json:"error_message"`
	DNSLookupMs     *float64  `json:"dns_lookup_ms"`
	TCPConnectMs    *float64  `json:"tcp_connect_ms"`
	TLSHandshakeMs  *float64  `json:"tls_handshake_ms"`
	TTFBMs          *float64  `json:"ttfb_ms"`
	TotalResponseMs *float64  `json:"total_response_ms"`
	SSLExpiryDays   *int      `json:"ssl_expiry_days"`
	ContentVerified bool      `json:"content_verified"`
}

// AlertRule represents condition rules for a service.
type AlertRule struct {
	ID        int       `json:"id"`
	ServiceID int       `json:"service_id"`
	Metric    string    `json:"metric"` // 'latency', 'status_code', 'ssl_expiry', 'content_verified'
	Operator  string    `json:"operator"` // '>', '<', '=', '!='
	Value     float64   `json:"value"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// AlertLog represents a triggered or resolved alert.
type AlertLog struct {
	ID          int        `json:"id"`
	ServiceID   int        `json:"service_id"`
	AlertRuleID *int       `json:"alert_rule_id"`
	TriggeredAt time.Time  `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	Status      string     `json:"status"` // 'active', 'resolved'
	Message     string     `json:"message"`
}
