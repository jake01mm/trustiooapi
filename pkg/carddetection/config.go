package carddetection

import (
	"time"
	
	"trusioo_api/config"
)

// NewConfigFromApp 从应用配置创建卡片检测配置
func NewConfigFromApp(appConfig *config.Config) *Config {
	if appConfig == nil || !appConfig.ThirdParty.CardDetectionEnabled {
		return nil
	}
	
	timeout := time.Duration(appConfig.ThirdParty.CardDetectionTimeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &Config{
		Host:      appConfig.ThirdParty.CardDetectionHost,
		AppID:     appConfig.ThirdParty.CardDetectionAppID,
		AppSecret: appConfig.ThirdParty.CardDetectionAppSecret,
		Timeout:   timeout,
	}
}

// NewConfigFromParams 从参数创建配置
func NewConfigFromParams(host, appID, appSecret string, timeout time.Duration) *Config {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &Config{
		Host:      host,
		AppID:     appID,
		AppSecret: appSecret,
		Timeout:   timeout,
	}
}

// IsEnabled 检查卡片检测是否启用
func IsEnabled(appConfig *config.Config) bool {
	return appConfig != nil && appConfig.ThirdParty.CardDetectionEnabled
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Host == "" {
		return ErrMissingHost
	}
	if c.AppID == "" {
		return ErrMissingAppID
	}
	if c.AppSecret == "" {
		return ErrMissingAppSecret
	}
	return nil
}