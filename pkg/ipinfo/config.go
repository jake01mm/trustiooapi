package ipinfo

import (
	"os"
	"strconv"
	"time"
)

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()
	
	if token := os.Getenv("IPINFO_TOKEN"); token != "" {
		config.Token = token
	}
	
	if baseURL := os.Getenv("IPINFO_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}
	
	if timeoutStr := os.Getenv("IPINFO_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.Timeout = timeout
		}
	}
	
	if maxRetriesStr := os.Getenv("IPINFO_MAX_RETRIES"); maxRetriesStr != "" {
		if maxRetries, err := strconv.Atoi(maxRetriesStr); err == nil {
			config.MaxRetries = maxRetries
		}
	}
	
	if retryDelayStr := os.Getenv("IPINFO_RETRY_DELAY"); retryDelayStr != "" {
		if retryDelay, err := time.ParseDuration(retryDelayStr); err == nil {
			config.RetryDelay = retryDelay
		}
	}
	
	if cacheEnableStr := os.Getenv("IPINFO_CACHE_ENABLE"); cacheEnableStr != "" {
		if cacheEnable, err := strconv.ParseBool(cacheEnableStr); err == nil {
			config.CacheEnable = cacheEnable
		}
	}
	
	if cacheTTLStr := os.Getenv("IPINFO_CACHE_TTL"); cacheTTLStr != "" {
		if cacheTTL, err := time.ParseDuration(cacheTTLStr); err == nil {
			config.CacheTTL = cacheTTL
		}
	}
	
	if maxConnsStr := os.Getenv("IPINFO_MAX_CONNS"); maxConnsStr != "" {
		if maxConns, err := strconv.Atoi(maxConnsStr); err == nil {
			config.MaxConns = maxConns
		}
	}
	
	if maxIdleConnsStr := os.Getenv("IPINFO_MAX_IDLE_CONNS"); maxIdleConnsStr != "" {
		if maxIdleConns, err := strconv.Atoi(maxIdleConnsStr); err == nil {
			config.MaxIdleConns = maxIdleConns
		}
	}
	
	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return NewError(ErrCodeInternal, "timeout must be positive", "")
	}
	
	if c.MaxRetries < 0 {
		return NewError(ErrCodeInternal, "max retries cannot be negative", "")
	}
	
	if c.RetryDelay < 0 {
		return NewError(ErrCodeInternal, "retry delay cannot be negative", "")
	}
	
	if c.CacheTTL <= 0 && c.CacheEnable {
		return NewError(ErrCodeInternal, "cache TTL must be positive when cache is enabled", "")
	}
	
	if c.MaxConns <= 0 {
		return NewError(ErrCodeInternal, "max connections must be positive", "")
	}
	
	if c.MaxIdleConns <= 0 {
		return NewError(ErrCodeInternal, "max idle connections must be positive", "")
	}
	
	if c.MaxIdleConns > c.MaxConns {
		return NewError(ErrCodeInternal, "max idle connections cannot exceed max connections", "")
	}
	
	return nil
}