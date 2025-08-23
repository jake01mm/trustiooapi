package carddetection

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock server responses
const (
	mockCheckCardSuccessResponse = `{"code":200,"msg":"","data":true}`
)

// Test configuration
var testConfig = &Config{
	Host:      "https://ckxiang.com",
	AppID:     "2508042205539611639",
	AppSecret: "2caa437312d44edcaf3ab61910cf31b7",
	Timeout:   10 * time.Second,
}

func TestNewClient(t *testing.T) {
	client := NewClient(testConfig)
	assert.NotNil(t, client)
	assert.Equal(t, testConfig, client.config)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.cryptoUtils)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr error
	}{
		{
			name:      "valid config",
			config:    testConfig,
			expectErr: nil,
		},
		{
			name: "missing host",
			config: &Config{
				AppID:     "test_app_id",
				AppSecret: "test_app_secret",
			},
			expectErr: ErrMissingHost,
		},
		{
			name: "missing app id",
			config: &Config{
				Host:      "http://test.example.com",
				AppSecret: "test_app_secret",
			},
			expectErr: ErrMissingAppID,
		},
		{
			name: "missing app secret",
			config: &Config{
				Host:  "http://test.example.com",
				AppID: "test_app_id",
			},
			expectErr: ErrMissingAppSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			err := client.ValidateConfig()
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckCard(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/userApiManage/checkCard", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test_app_id", r.Header.Get("appId"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockCheckCardSuccessResponse))
	}))
	defer server.Close()

	config := &Config{
		Host:      server.URL,
		AppID:     "test_app_id",
		AppSecret: "test_app_secret",
		Timeout:   10 * time.Second,
	}

	client := NewClient(config)
	ctx := context.Background()

	req := &CheckCardRequest{
		Cards:       []string{"X123123123123123"},
		ProductMark: ProductMarkItunes,
		RegionID:    2,
		RegionName:  "美国",
		AutoType:    0,
	}

	resp, err := client.CheckCard(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.Code)
	assert.True(t, resp.Data)
}

func TestCheckCardValidation(t *testing.T) {
	client := NewClient(testConfig)
	ctx := context.Background()

	tests := []struct {
		name      string
		request   *CheckCardRequest
		expectErr bool
	}{
		{
			name: "valid iTunes request",
			request: &CheckCardRequest{
				Cards:       []string{"X123123123123123"},
				ProductMark: ProductMarkItunes,
				RegionID:    2,
				RegionName:  "美国",
			},
			expectErr: false,
		},
		{
			name: "empty cards",
			request: &CheckCardRequest{
				Cards:       []string{},
				ProductMark: ProductMarkItunes,
				RegionID:    2,
			},
			expectErr: true,
		},
		{
			name: "missing product mark",
			request: &CheckCardRequest{
				Cards: []string{"X123123123123123"},
			},
			expectErr: true,
		},
		{
			name: "iTunes without region",
			request: &CheckCardRequest{
				Cards:       []string{"X123123123123123"},
				ProductMark: ProductMarkItunes,
			},
			expectErr: true,
		},
		{
			name: "Amazon without region",
			request: &CheckCardRequest{
				Cards:       []string{"123123123123123"},
				ProductMark: ProductMarkAmazon,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CheckCard(ctx, tt.request)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				// Note: This will fail due to network call, but validation should pass
				if err != nil {
					assert.Contains(t, err.Error(), "request failed")
				}
			}
		})
	}
}

func TestCheckCardResult(t *testing.T) {
	// Create a test result that we'll encrypt
	testResult := CardResult{
		CardNo:     "X123123123123123",
		Status:     CardStatusValid,
		PinCode:    "",
		Message:    "",
		CheckTime:  "2024-01-01 12:00:00",
		RegionName: "美国",
		RegionID:   2,
	}

	resultData, _ := json.Marshal(testResult)

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/userApiManage/checkCardResult", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test_app_id", r.Header.Get("appId"))

		// Create crypto utils to encrypt the result
		crypto := NewCryptoUtils("test_app_secret")
		encryptedResult, _ := crypto.DESEncrypt(string(resultData))

		response := CheckCardResultResponse{
			Code: 200,
			Msg:  "",
			Data: encryptedResult,
		}

		respData, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respData)
	}))
	defer server.Close()

	config := &Config{
		Host:      server.URL,
		AppID:     "test_app_id",
		AppSecret: "test_app_secret",
		Timeout:   10 * time.Second,
	}

	client := NewClient(config)
	ctx := context.Background()

	req := &CheckCardResultRequest{
		ProductMark: ProductMarkItunes,
		CardNo:      "X123123123123123",
		PinCode:     "",
	}

	result, err := client.CheckCardResult(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "X123123123123123", result.CardNo)
	assert.Equal(t, CardStatusValid, result.Status)
}

func TestCryptoUtils(t *testing.T) {
	appSecret := "test_secret"
	crypto := NewCryptoUtils(appSecret)

	t.Run("Sign and Verify", func(t *testing.T) {
		params := map[string]interface{}{
			"cardNo":      "X123123123123123",
			"productMark": "iTunes",
			"timestamp":   "1234567890",
		}

		sign, err := crypto.Sign(params)
		require.NoError(t, err)
		assert.NotEmpty(t, sign)

		// Add sign to params and verify
		params["sign"] = sign
		assert.True(t, crypto.VerifySign(params, sign))

		// Test with wrong sign
		assert.False(t, crypto.VerifySign(params, "wrong_sign"))
	})

	t.Run("Encrypt and Decrypt", func(t *testing.T) {
		plainText := `{"cardNo":"X123123123123123","productMark":"iTunes","timestamp":"1234567890"}`

		encrypted, err := crypto.DESEncrypt(plainText)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, plainText, encrypted)

		decrypted, err := crypto.DESDecrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plainText, decrypted)
	})
}

func TestProductMarkValidation(t *testing.T) {
	tests := []struct {
		name        string
		productMark ProductMark
		valid       bool
	}{
		{"iTunes", ProductMarkItunes, true},
		{"Amazon", ProductMarkAmazon, true},
		{"Xbox", ProductMarkXbox, true},
		{"Nike", ProductMarkNike, true},
		{"Sephora", ProductMarkSephora, true},
		{"Razer", ProductMarkRazer, true},
		{"ND", ProductMarkND, true},
		{"Invalid", ProductMark("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			validMarks := []ProductMark{
				ProductMarkItunes, ProductMarkAmazon, ProductMarkXbox,
				ProductMarkNike, ProductMarkSephora, ProductMarkRazer, ProductMarkND,
			}
			for _, mark := range validMarks {
				if tt.productMark == mark {
					found = true
					break
				}
			}
			assert.Equal(t, tt.valid, found)
		})
	}
}

func TestRegionValidation(t *testing.T) {
	client := NewClient(testConfig)

	t.Run("iTunes regions", func(t *testing.T) {
		assert.True(t, client.isValidITunesRegion(1))   // 英国
		assert.True(t, client.isValidITunesRegion(2))   // 美国
		assert.False(t, client.isValidITunesRegion(99)) // 不存在
	})

	t.Run("Amazon regions", func(t *testing.T) {
		assert.True(t, client.isValidAmazonRegion(1))   // 欧盟区
		assert.True(t, client.isValidAmazonRegion(2))   // 美亚/加亚
		assert.False(t, client.isValidAmazonRegion(99)) // 不存在
	})

	t.Run("Xbox regions", func(t *testing.T) {
		assert.True(t, client.isValidXboxRegion("美国"))
		assert.True(t, client.isValidXboxRegion("加拿大"))
		assert.False(t, client.isValidXboxRegion("不存在的地区"))
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("CardDetectionError", func(t *testing.T) {
		err := NewError(ErrCodeInvalidRequest, "test error", nil)
		assert.True(t, IsCardDetectionError(err))
		assert.Equal(t, ErrCodeInvalidRequest, GetErrorCode(err))
		assert.Contains(t, err.Error(), "test error")
	})

	t.Run("Wrapped error", func(t *testing.T) {
		originalErr := assert.AnError
		wrappedErr := WrapError(originalErr, ErrCodeAPIRequest, "wrapped error")
		assert.True(t, IsCardDetectionError(wrappedErr))
		assert.Equal(t, ErrCodeAPIRequest, GetErrorCode(wrappedErr))
		assert.Equal(t, originalErr, wrappedErr.Unwrap())
	})
}

func TestConfigHelpers(t *testing.T) {
	t.Run("Config validation", func(t *testing.T) {
		config := &Config{
			Host:      "http://test.example.com",
			AppID:     "test_id",
			AppSecret: "test_secret",
		}
		assert.NoError(t, config.Validate())

		config.Host = ""
		assert.Equal(t, ErrMissingHost, config.Validate())
	})

	t.Run("NewConfigFromParams", func(t *testing.T) {
		config := NewConfigFromParams("host", "id", "secret", 30*time.Second)
		assert.Equal(t, "host", config.Host)
		assert.Equal(t, "id", config.AppID)
		assert.Equal(t, "secret", config.AppSecret)
		assert.Equal(t, 30*time.Second, config.Timeout)
	})
}
