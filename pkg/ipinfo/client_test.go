package ipinfo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Token:       "test-token",
		Timeout:     5 * time.Second,
		CacheEnable: true,
	}
	
	client := NewClient(config)
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	
	defer client.Close()
}

func TestGetIPInfo(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"ip": "8.8.8.8",
			"city": "Mountain View",
			"region": "California",
			"country": "US",
			"loc": "37.4056,-122.0775",
			"org": "AS15169 Google LLC",
			"timezone": "America/Los_Angeles"
		}`))
	}))
	defer server.Close()
	
	config := &Config{
		BaseURL:     server.URL,
		Timeout:     5 * time.Second,
		CacheEnable: false,
	}
	
	client := NewClient(config)
	defer client.Close()
	
	ctx := context.Background()
	info, err := client.GetIPInfo(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if info.IP != "8.8.8.8" {
		t.Errorf("Expected IP 8.8.8.8, got %s", info.IP)
	}
	
	if info.City != "Mountain View" {
		t.Errorf("Expected city Mountain View, got %s", info.City)
	}
}

func TestGetIPInfoWithCache(t *testing.T) {
	requestCount := 0
	
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"ip": "8.8.8.8",
			"city": "Mountain View",
			"region": "California",
			"country": "US"
		}`))
	}))
	defer server.Close()
	
	config := &Config{
		BaseURL:     server.URL,
		Timeout:     5 * time.Second,
		CacheEnable: true,
		CacheTTL:    1 * time.Minute,
	}
	
	client := NewClient(config)
	defer client.Close()
	
	ctx := context.Background()
	
	// First request
	_, err := client.GetIPInfo(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Second request should use cache
	_, err = client.GetIPInfo(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Should only have made one HTTP request
	if requestCount != 1 {
		t.Errorf("Expected 1 HTTP request, got %d", requestCount)
	}
}

func TestBatchGetIPInfo(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Return different data based on IP
		if r.URL.Path == "/8.8.8.8" {
			w.Write([]byte(`{"ip": "8.8.8.8", "city": "Mountain View"}`))
		} else if r.URL.Path == "/1.1.1.1" {
			w.Write([]byte(`{"ip": "1.1.1.1", "city": "San Francisco"}`))
		}
	}))
	defer server.Close()
	
	config := &Config{
		BaseURL:     server.URL,
		Timeout:     5 * time.Second,
		CacheEnable: false,
	}
	
	client := NewClient(config)
	defer client.Close()
	
	ctx := context.Background()
	ips := []string{"8.8.8.8", "1.1.1.1"}
	
	results, err := client.BatchGetIPInfo(ctx, ips)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	if results["8.8.8.8"] == nil {
		t.Error("Expected result for 8.8.8.8")
	}
	
	if results["1.1.1.1"] == nil {
		t.Error("Expected result for 1.1.1.1")
	}
}

func TestValidateIP(t *testing.T) {
	client := &client{}
	
	// Valid IPs
	validIPs := []string{"8.8.8.8", "192.168.1.1", "::1", "2001:db8::1"}
	for _, ip := range validIPs {
		if err := client.validateIP(ip); err != nil {
			t.Errorf("Expected %s to be valid, got error: %v", ip, err)
		}
	}
	
	// Invalid IPs
	invalidIPs := []string{"", "invalid", "256.256.256.256", "8.8.8"}
	for _, ip := range invalidIPs {
		if err := client.validateIP(ip); err == nil {
			t.Errorf("Expected %s to be invalid", ip)
		}
	}
}

func TestConfigValidation(t *testing.T) {
	// Valid config
	validConfig := &Config{
		Timeout:      5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   time.Second,
		CacheEnable:  true,
		CacheTTL:     10 * time.Minute,
		MaxConns:     100,
		MaxIdleConns: 50,
	}
	
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Expected valid config to pass validation, got: %v", err)
	}
	
	// Invalid configs
	invalidConfigs := []*Config{
		{Timeout: -1 * time.Second},                           // Negative timeout
		{Timeout: time.Second, MaxRetries: -1},                // Negative retries
		{Timeout: time.Second, RetryDelay: -1 * time.Second},  // Negative retry delay
		{Timeout: time.Second, MaxConns: -1},                  // Negative max connections
		{Timeout: time.Second, MaxConns: 10, MaxIdleConns: 20}, // Idle > max
	}
	
	for i, config := range invalidConfigs {
		if err := config.Validate(); err == nil {
			t.Errorf("Expected invalid config %d to fail validation", i)
		}
	}
}