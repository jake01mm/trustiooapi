package ipinfo

import (
	"context"
	"fmt"
	"time"
)

// IPInfo represents the response from ipinfo.io API
type IPInfo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	ASN      *ASN   `json:"asn,omitempty"`
	Company  *Company `json:"company,omitempty"`
	Privacy  *Privacy `json:"privacy,omitempty"`
	Abuse    *Abuse   `json:"abuse,omitempty"`
	Domains  *Domains `json:"domains,omitempty"`
}

// ASN represents autonomous system number information
type ASN struct {
	ASN    string `json:"asn"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Route  string `json:"route"`
	Type   string `json:"type"`
}

// Company represents company information
type Company struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Type   string `json:"type"`
}

// Privacy represents privacy information
type Privacy struct {
	VPN    bool `json:"vpn"`
	Proxy  bool `json:"proxy"`
	Tor    bool `json:"tor"`
	Relay  bool `json:"relay"`
	Hosting bool `json:"hosting"`
	Service string `json:"service"`
}

// Abuse represents abuse contact information
type Abuse struct {
	Address string `json:"address"`
	Country string `json:"country"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Network string `json:"network"`
	Phone   string `json:"phone"`
}

// Domains represents domains information
type Domains struct {
	IP      string   `json:"ip"`
	Page    int      `json:"page"`
	Total   int      `json:"total"`
	Domains []string `json:"domains"`
}

// Config represents the configuration for IPInfo client
type Config struct {
	Token       string        `json:"token"`
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
	CacheEnable bool          `json:"cache_enable"`
	CacheTTL    time.Duration `json:"cache_ttl"`
	MaxConns    int           `json:"max_conns"`
	MaxIdleConns int          `json:"max_idle_conns"`
}

// Client interface for IPInfo operations
type Client interface {
	GetIPInfo(ctx context.Context, ip string) (*IPInfo, error)
	GetMyIP(ctx context.Context) (*IPInfo, error)
	BatchGetIPInfo(ctx context.Context, ips []string) (map[string]*IPInfo, error)
	Close() error
}

// Cache interface for caching operations
type Cache interface {
	Get(key string) (*IPInfo, bool)
	Set(key string, value *IPInfo, ttl time.Duration)
	Delete(key string)
	Clear()
}

// ErrorCode represents error codes
type ErrorCode int

const (
	ErrCodeInvalidIP ErrorCode = iota + 1000
	ErrCodeAPIRequest
	ErrCodeAPIResponse
	ErrCodeTimeout
	ErrCodeRateLimit
	ErrCodeUnauthorized
	ErrCodeNotFound
	ErrCodeInternal
)

// Error represents an IPInfo error
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	IP      string    `json:"ip,omitempty"`
}

func (e *Error) Error() string {
	if e.IP != "" {
		return fmt.Sprintf("ipinfo error [%d]: %s (IP: %s)", e.Code, e.Message, e.IP)
	}
	return fmt.Sprintf("ipinfo error [%d]: %s", e.Code, e.Message)
}

// NewError creates a new IPInfo error
func NewError(code ErrorCode, message string, ip string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		IP:      ip,
	}
}