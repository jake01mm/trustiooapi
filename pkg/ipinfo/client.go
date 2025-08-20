package ipinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// client implements the Client interface
type client struct {
	config     *Config
	httpClient *http.Client
	cache      Cache
	limiter    chan struct{}
	closeOnce  sync.Once
}

// NewClient creates a new IPInfo client
func NewClient(config *Config) Client {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Set defaults if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://ipinfo.io"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 30 * time.Minute
	}
	if config.MaxConns == 0 {
		config.MaxConns = 100
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 50
	}
	
	// Create HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConns / 2,
		MaxConnsPerHost:     config.MaxConns,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
	
	var cache Cache
	if config.CacheEnable {
		cache = NewMemoryCache()
	}
	
	// Create rate limiter channel (100 requests per second max)
	limiter := make(chan struct{}, 100)
	
	return &client{
		config:     config,
		httpClient: httpClient,
		cache:      cache,
		limiter:    limiter,
	}
}

// GetIPInfo retrieves information for a specific IP address
func (c *client) GetIPInfo(ctx context.Context, ip string) (*IPInfo, error) {
	if err := c.validateIP(ip); err != nil {
		return nil, err
	}
	
	// Check cache first
	if c.cache != nil {
		if cached, found := c.cache.Get(ip); found {
			return cached, nil
		}
	}
	
	// Make API request
	url := c.buildURL(ip)
	info, err := c.makeRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	if c.cache != nil {
		c.cache.Set(ip, info, c.config.CacheTTL)
	}
	
	return info, nil
}

// GetMyIP retrieves information for the current IP address
func (c *client) GetMyIP(ctx context.Context) (*IPInfo, error) {
	url := c.buildURL("")
	return c.makeRequest(ctx, url)
}

// BatchGetIPInfo retrieves information for multiple IP addresses concurrently
func (c *client) BatchGetIPInfo(ctx context.Context, ips []string) (map[string]*IPInfo, error) {
	if len(ips) == 0 {
		return nil, NewError(ErrCodeInvalidIP, "no IPs provided", "")
	}
	
	results := make(map[string]*IPInfo)
	resultsMutex := sync.Mutex{}
	
	// Use semaphore to limit concurrent requests
	semaphore := make(chan struct{}, 10) // Max 10 concurrent requests
	var wg sync.WaitGroup
	
	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			info, err := c.GetIPInfo(ctx, ip)
			
			resultsMutex.Lock()
			if err == nil {
				results[ip] = info
			}
			resultsMutex.Unlock()
		}(ip)
	}
	
	wg.Wait()
	return results, nil
}

// Close closes the client and releases resources
func (c *client) Close() error {
	c.closeOnce.Do(func() {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
		if c.cache != nil {
			c.cache.Clear()
		}
		close(c.limiter)
	})
	return nil
}

// validateIP validates the IP address format
func (c *client) validateIP(ip string) error {
	if ip == "" {
		return NewError(ErrCodeInvalidIP, "IP address cannot be empty", ip)
	}
	
	if net.ParseIP(ip) == nil {
		return NewError(ErrCodeInvalidIP, "invalid IP address format", ip)
	}
	
	return nil
}

// buildURL constructs the API URL
func (c *client) buildURL(ip string) string {
	baseURL := strings.TrimRight(c.config.BaseURL, "/")
	
	if ip == "" {
		// Get current IP
		if c.config.Token != "" {
			return fmt.Sprintf("%s?token=%s", baseURL, c.config.Token)
		}
		return baseURL
	}
	
	// Get specific IP info
	if c.config.Token != "" {
		return fmt.Sprintf("%s/%s?token=%s", baseURL, ip, c.config.Token)
	}
	return fmt.Sprintf("%s/%s", baseURL, ip)
}

// makeRequest makes the HTTP request with retries
func (c *client) makeRequest(ctx context.Context, url string) (*IPInfo, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, NewError(ErrCodeTimeout, "context canceled during retry", "")
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			}
		}
		
		// Rate limiting
		select {
		case c.limiter <- struct{}{}:
			defer func() { <-c.limiter }()
		case <-ctx.Done():
			return nil, NewError(ErrCodeTimeout, "context canceled waiting for rate limit", "")
		}
		
		info, err := c.doRequest(ctx, url)
		if err == nil {
			return info, nil
		}
		
		lastErr = err
		
		// Don't retry on certain errors
		if ipErr, ok := err.(*Error); ok {
			switch ipErr.Code {
			case ErrCodeInvalidIP, ErrCodeUnauthorized, ErrCodeNotFound:
				return nil, err
			}
		}
	}
	
	return nil, lastErr
}

// doRequest performs the actual HTTP request
func (c *client) doRequest(ctx context.Context, url string) (*IPInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, NewError(ErrCodeAPIRequest, "failed to create request", "")
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-ipinfo-client/1.0")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, NewError(ErrCodeTimeout, "request timeout", "")
		}
		return nil, NewError(ErrCodeAPIRequest, "HTTP request failed: "+err.Error(), "")
	}
	defer resp.Body.Close()
	
	// Check status code
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusUnauthorized:
		return nil, NewError(ErrCodeUnauthorized, "unauthorized - check your token", "")
	case http.StatusNotFound:
		return nil, NewError(ErrCodeNotFound, "IP not found", "")
	case http.StatusTooManyRequests:
		return nil, NewError(ErrCodeRateLimit, "rate limit exceeded", "")
	default:
		return nil, NewError(ErrCodeAPIResponse, fmt.Sprintf("API returned status %d", resp.StatusCode), "")
	}
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewError(ErrCodeAPIResponse, "failed to read response body", "")
	}
	
	// Parse JSON response
	var info IPInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, NewError(ErrCodeAPIResponse, "failed to parse JSON response", "")
	}
	
	return &info, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:      "https://ipinfo.io",
		Timeout:      10 * time.Second,
		MaxRetries:   3,
		RetryDelay:   time.Second,
		CacheEnable:  true,
		CacheTTL:     30 * time.Minute,
		MaxConns:     100,
		MaxIdleConns: 50,
	}
}