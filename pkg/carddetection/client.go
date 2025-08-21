package carddetection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client 卡片检测客户端
type Client struct {
	config      *Config
	httpClient  *http.Client
	cryptoUtils *CryptoUtils
}

// NewClient 创建新的客户端
func NewClient(config *Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		cryptoUtils: NewCryptoUtils(config.AppSecret),
	}
}

// ValidateConfig 验证配置
func (c *Client) ValidateConfig() error {
	if c.config.Host == "" {
		return ErrMissingHost
	}
	if c.config.AppID == "" {
		return ErrMissingAppID
	}
	if c.config.AppSecret == "" {
		return ErrMissingAppSecret
	}
	return nil
}

// CheckCard 执行卡片检测
func (c *Client) CheckCard(ctx context.Context, req *CheckCardRequest) (*CheckCardResponse, error) {
	if err := c.ValidateConfig(); err != nil {
		return nil, err
	}
	
	if err := c.validateCheckCardRequest(req); err != nil {
		return nil, err
	}
	
	// 构建内部请求数据
	internalReq := &internalRequest{
		Cards:       req.Cards,
		ProductMark: req.ProductMark,
		RegionID:    req.RegionID,
		RegionName:  req.RegionName,
		AutoType:    req.AutoType,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	
	// 生成签名
	if err := c.signRequest(internalReq); err != nil {
		return nil, WrapError(err, ErrCodeSignatureFailed, "failed to sign request")
	}
	
	// 加密请求数据
	encryptedData, err := c.encryptRequest(internalReq)
	if err != nil {
		return nil, WrapError(err, ErrCodeEncryptionFailed, "failed to encrypt request")
	}
	
	// 发送HTTP请求
	url := fmt.Sprintf("%s/api/userApiManage/checkCard", c.config.Host)
	resp, err := c.sendRequest(ctx, url, encryptedData)
	if err != nil {
		return nil, err
	}
	
	return resp.(*CheckCardResponse), nil
}

// CheckCardResult 查询卡片检测结果
func (c *Client) CheckCardResult(ctx context.Context, req *CheckCardResultRequest) (*CardResult, error) {
	if err := c.ValidateConfig(); err != nil {
		return nil, err
	}
	
	if err := c.validateCheckCardResultRequest(req); err != nil {
		return nil, err
	}
	
	// 构建内部请求数据
	internalReq := &internalRequest{
		CardNo:      req.CardNo,
		PinCode:     req.PinCode,
		ProductMark: req.ProductMark,
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
	}
	
	// 生成签名
	if err := c.signRequest(internalReq); err != nil {
		return nil, WrapError(err, ErrCodeSignatureFailed, "failed to sign request")
	}
	
	// 加密请求数据
	encryptedData, err := c.encryptRequest(internalReq)
	if err != nil {
		return nil, WrapError(err, ErrCodeEncryptionFailed, "failed to encrypt request")
	}
	
	// 发送HTTP请求
	url := fmt.Sprintf("%s/api/userApiManage/checkCardResult", c.config.Host)
	resp, err := c.sendRequest(ctx, url, encryptedData)
	if err != nil {
		return nil, err
	}
	
	// 解密响应数据
	resultResp := resp.(*CheckCardResultResponse)
	if resultResp.Code != 200 {
		return nil, NewError(ErrCodeAPIResponse, resultResp.Msg, nil)
	}
	
	// 解密结果数据
	decryptedData, err := c.cryptoUtils.DESDecrypt(resultResp.Data)
	if err != nil {
		return nil, WrapError(err, ErrCodeDecryptionFailed, "failed to decrypt response data")
	}
	
	// 解析结果
	var result CardResult
	if parseErr := json.Unmarshal([]byte(decryptedData), &result); parseErr != nil {
		return nil, WrapError(parseErr, ErrCodeAPIResponse, "failed to parse decrypted result")
	}
	
	return &result, nil
}

// signRequest 为请求生成签名
func (c *Client) signRequest(req *internalRequest) error {
	// 手动构建参数map，确保与Java版本一致
	params := make(map[string]interface{})
	
	// 只添加非空值
	if len(req.Cards) > 0 {
		params["cards"] = req.Cards
	}
	if req.CardNo != "" {
		params["cardNo"] = req.CardNo
	}
	if req.PinCode != "" {
		params["pinCode"] = req.PinCode
	}
	if req.ProductMark != "" {
		params["productMark"] = string(req.ProductMark)
	}
	if req.RegionID != 0 {
		params["regionId"] = req.RegionID
	}
	if req.RegionName != "" {
		params["regionName"] = req.RegionName
	}
	if req.AutoType != 0 {
		params["autoType"] = req.AutoType
	}
	if req.Timestamp != "" {
		params["timestamp"] = req.Timestamp
	}
	
	sign, err := c.cryptoUtils.Sign(params)
	if err != nil {
		return err
	}
	
	req.Sign = sign
	return nil
}

// encryptRequest 加密请求数据
func (c *Client) encryptRequest(req *internalRequest) (*encryptedRequest, error) {
	// 序列化内部请求
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	// 加密数据
	encryptedData, err := c.cryptoUtils.DESEncrypt(string(data))
	if err != nil {
		return nil, err
	}
	
	return &encryptedRequest{
		Data: encryptedData,
	}, nil
}

// sendRequest 发送HTTP请求
func (c *Client) sendRequest(ctx context.Context, url string, encReq *encryptedRequest) (interface{}, error) {
	// 序列化请求体
	reqBody, err := json.Marshal(encReq)
	if err != nil {
		return nil, WrapError(err, ErrCodeAPIRequest, "failed to marshal request")
	}
	
	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, WrapError(err, ErrCodeAPIRequest, "failed to create HTTP request")
	}
	
	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("appId", c.config.AppID)
	
	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, WrapError(err, ErrCodeTimeout, "request timeout")
		}
		return nil, WrapError(err, ErrCodeAPIRequest, "HTTP request failed")
	}
	defer resp.Body.Close()
	
	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, WrapError(err, ErrCodeAPIResponse, "failed to read response body")
	}
	
	// 根据URL判断响应类型并解析
	if contains(url, "checkCardResult") {
		var result CheckCardResultResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, WrapError(err, ErrCodeAPIResponse, "failed to parse response")
		}
		return &result, nil
	} else {
		var result CheckCardResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, WrapError(err, ErrCodeAPIResponse, "failed to parse response")
		}
		return &result, nil
	}
}

// validateCheckCardRequest 验证测卡请求
func (c *Client) validateCheckCardRequest(req *CheckCardRequest) error {
	if len(req.Cards) == 0 {
		return NewError(ErrCodeInvalidRequest, "cards cannot be empty", nil)
	}
	
	if req.ProductMark == "" {
		return ErrInvalidProductMark
	}
	
	// 验证产品类型和地区的组合
	switch req.ProductMark {
	case ProductMarkItunes:
		if req.RegionID == 0 && req.AutoType == 0 {
			return NewError(ErrCodeInvalidRequest, "iTunes cards require regionId or autoType=1", nil)
		}
		if req.RegionID != 0 && !c.isValidITunesRegion(req.RegionID) {
			return ErrUnsupportedRegion
		}
	case ProductMarkAmazon:
		if req.RegionID == 0 {
			return NewError(ErrCodeInvalidRequest, "Amazon cards require regionId", nil)
		}
		if !c.isValidAmazonRegion(req.RegionID) {
			return ErrUnsupportedRegion
		}
	case ProductMarkRazer:
		if req.RegionID == 0 {
			return NewError(ErrCodeInvalidRequest, "Razer cards require regionId", nil)
		}
		if !c.isValidRazerRegion(req.RegionID) {
			return ErrUnsupportedRegion
		}
	case ProductMarkXbox:
		if req.RegionName == "" {
			return NewError(ErrCodeInvalidRequest, "Xbox cards require regionName", nil)
		}
		if !c.isValidXboxRegion(req.RegionName) {
			return ErrUnsupportedRegion
		}
	}
	
	return nil
}

// validateCheckCardResultRequest 验证查询结果请求
func (c *Client) validateCheckCardResultRequest(req *CheckCardResultRequest) error {
	if req.CardNo == "" {
		return NewError(ErrCodeInvalidRequest, "cardNo cannot be empty", nil)
	}
	
	if req.ProductMark == "" {
		return ErrInvalidProductMark
	}
	
	// 某些产品需要PIN码
	if (req.ProductMark == ProductMarkSephora || req.ProductMark == ProductMarkNike || req.ProductMark == ProductMarkND) && req.PinCode == "" {
		return NewError(ErrCodeInvalidRequest, fmt.Sprintf("%s cards require pinCode", req.ProductMark), nil)
	}
	
	return nil
}

// 地区验证辅助函数
func (c *Client) isValidITunesRegion(regionID int) bool {
	for _, region := range ITunesRegions {
		if region.ID == regionID {
			return true
		}
	}
	return false
}

func (c *Client) isValidAmazonRegion(regionID int) bool {
	for _, region := range AmazonRegions {
		if region.ID == regionID {
			return true
		}
	}
	return false
}

func (c *Client) isValidRazerRegion(regionID int) bool {
	for _, region := range RazerRegions {
		if region.ID == regionID {
			return true
		}
	}
	return false
}

func (c *Client) isValidXboxRegion(regionName string) bool {
	for _, region := range XboxRegions {
		if region == regionName {
			return true
		}
	}
	return false
}

// 辅助函数
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}