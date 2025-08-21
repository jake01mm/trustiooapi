package carddetection

import (
	"fmt"
	"time"
)

// ProductMark 产品类型枚举
type ProductMark string

const (
	ProductMarkSephora ProductMark = "sephora"  // 丝芙兰
	ProductMarkRazer   ProductMark = "Razer"    // 雷蛇
	ProductMarkItunes  ProductMark = "iTunes"   // 苹果
	ProductMarkAmazon  ProductMark = "amazon"   // 亚马逊
	ProductMarkXbox    ProductMark = "xBox"     // XBOX
	ProductMarkNike    ProductMark = "nike"     // NIKE
	ProductMarkND      ProductMark = "nd"       // ND
)

// CardStatus 卡片状态
type CardStatus int

const (
	CardStatusWaiting   CardStatus = 0 // 等待检测
	CardStatusTesting   CardStatus = 1 // 测卡中
	CardStatusValid     CardStatus = 2 // 有效
	CardStatusInvalid   CardStatus = 3 // 无效
	CardStatusRedeemed  CardStatus = 4 // 已兑换
	CardStatusFailed    CardStatus = 5 // 检测失败
	CardStatusLowPoints CardStatus = 6 // 点数不足
)

// CheckCardRequest 测卡请求
type CheckCardRequest struct {
	Cards       []string    `json:"cards" binding:"required"`       // 卡号列表
	ProductMark ProductMark `json:"productMark" binding:"required"` // 产品类型
	RegionID    int         `json:"regionId,omitempty"`             // 地区ID（部分产品需要）
	RegionName  string      `json:"regionName,omitempty"`           // 地区名称（部分产品需要）
	AutoType    int         `json:"autoType,omitempty"`             // 苹果测卡专用：0指定国家 1自动识别
}

// CheckCardResponse 测卡响应
type CheckCardResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}

// CheckCardResultRequest 查询测卡结果请求
type CheckCardResultRequest struct {
	ProductMark ProductMark `json:"productMark" binding:"required"` // 产品类型
	CardNo      string      `json:"cardNo" binding:"required"`      // 卡号
	PinCode     string      `json:"pinCode,omitempty"`              // PIN码（某些卡片需要）
}

// CheckCardResultResponse 查询测卡结果响应
type CheckCardResultResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"` // AES加密的结果，需要解密
}

// CardResult 卡片检测结果（解密后的数据）
type CardResult struct {
	CardNo     string     `json:"cardNo"`     // 请求的卡号
	Status     CardStatus `json:"status"`     // 状态
	PinCode    string     `json:"pinCode"`    // PIN码
	Message    string     `json:"message"`    // 检测结果信息（错误信息）
	CheckTime  interface{} `json:"checkTime"`  // 检测时间（可能是字符串或数字）
	RegionName string     `json:"regionName"` // 卡种国家（部分含有）
	RegionID   int        `json:"regionId"`   // 卡种国家编号（部分含有）
}

// GetCheckTimeString 获取格式化的检测时间字符串
func (r *CardResult) GetCheckTimeString() string {
	switch v := r.CheckTime.(type) {
	case string:
		return v
	case float64:
		// 如果是Unix时间戳，转换为时间字符串
		if v > 1000000000 { // 合理的Unix时间戳范围
			return time.Unix(int64(v), 0).Format("2006-01-02 15:04:05")
		}
		return fmt.Sprintf("%.0f", v)
	case int64:
		if v > 1000000000 {
			return time.Unix(v, 0).Format("2006-01-02 15:04:05")
		}
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// RegionInfo 地区信息
type RegionInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 各产品支持的地区配置
var (
	// iTunes支持的地区
	ITunesRegions = []RegionInfo{
		{1, "英国"}, {2, "美国"}, {3, "德国"}, {4, "澳大利亚"},
		{5, "加拿大"}, {6, "日本"}, {8, "西班牙"}, {9, "意大利"},
		{10, "法国"}, {11, "爱尔兰"}, {12, "墨西哥"},
	}

	// Amazon支持的地区
	AmazonRegions = []RegionInfo{
		{2, "美亚/加亚"}, {1, "欧盟区"},
	}

	// Razer支持的地区
	RazerRegions = []RegionInfo{
		{12, "美国"}, {6, "澳大利亚"}, {13, "巴西"}, {26, "柬埔寨"},
		{20, "加拿大"}, {25, "智利"}, {22, "哥伦比亚"}, {17, "香港特别行政区"},
		{4, "印度"}, {7, "印度尼西亚"}, {27, "日本"}, {1, "马来西亚"},
		{19, "缅甸"}, {15, "新西兰"}, {29, "巴基斯坦"}, {8, "菲律宾"},
		{5, "新加坡"}, {18, "土耳其"}, {33, "越南"}, {2, "其他"},
		{28, "其他（中文）"}, {21, "墨西哥"},
	}

	// Xbox支持的地区
	XboxRegions = []string{
		"美国", "加拿大", "英国", "澳大利亚", "新西兰", "新加坡",
		"韩国", "墨西哥", "瑞典", "哥伦比亚", "阿根廷", "尼日利亚",
		"香港特别行政区", "挪威", "波兰", "德国",
	}
)

// internalRequest 内部请求数据结构（用于加密）
type internalRequest struct {
	Cards       []string    `json:"cards,omitempty"`
	CardNo      string      `json:"cardNo,omitempty"`
	PinCode     string      `json:"pinCode,omitempty"`
	ProductMark ProductMark `json:"productMark"`
	RegionID    int         `json:"regionId,omitempty"`
	RegionName  string      `json:"regionName,omitempty"`
	AutoType    int         `json:"autoType,omitempty"`
	Timestamp   string      `json:"timestamp"`
	Sign        string      `json:"sign"`
}

// encryptedRequest API请求数据结构
type encryptedRequest struct {
	Data string `json:"data"`
}

// Config 卡片检测配置
type Config struct {
	Host      string        // API主机地址
	AppID     string        // 应用ID
	AppSecret string        // 应用密钥
	Timeout   time.Duration // 请求超时时间
}