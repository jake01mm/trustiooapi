package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"
)

// 性能测试结果
type TestResult struct {
	TestName        string        `json:"test_name"`
	TotalRequests   int           `json:"total_requests"`
	SuccessRequests int           `json:"success_requests"`
	FailedRequests  int           `json:"failed_requests"`
	AverageLatency  time.Duration `json:"average_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	RequestsPerSec  float64       `json:"requests_per_second"`
	ErrorRate       float64       `json:"error_rate"`
	Duration        time.Duration `json:"duration"`
}

// 测试配置
type TestConfig struct {
	BaseURL         string
	Concurrency     int
	TotalRequests   int
	TestDuration    time.Duration
	AuthToken       string
	TestImagePath   string
}

// 性能测试器
type PerformanceTester struct {
	config TestConfig
	client *http.Client
}

func NewPerformanceTester(config TestConfig) *PerformanceTester {
	return &PerformanceTester{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func main() {
	fmt.Println("🚀 Starting Trusioo API Performance Tests...")
	
	// 检查环境变量
	baseURL := getEnv("API_BASE_URL", "http://localhost:8080")
	authToken := getEnv("AUTH_TOKEN", "")
	testImagePath := getEnv("TEST_IMAGE_PATH", "test_image.jpg")
	
	if authToken == "" {
		fmt.Println("⚠️  Warning: No AUTH_TOKEN provided. Some tests may fail.")
	}
	
	config := TestConfig{
		BaseURL:         baseURL,
		Concurrency:     10,
		TotalRequests:   1000,
		TestDuration:    5 * time.Minute,
		AuthToken:       authToken,
		TestImagePath:   testImagePath,
	}
	
	tester := NewPerformanceTester(config)
	results := []TestResult{}
	
	// 1. API健康检查测试
	fmt.Println("\n📊 Running Health Check Load Test...")
	result := tester.RunHealthCheckTest()
	results = append(results, result)
	printTestResult(result)
	
	// 2. 图片列表查询测试
	fmt.Println("\n📊 Running Image List Query Test...")
	result = tester.RunImageListTest()
	results = append(results, result)
	printTestResult(result)
	
	// 3. 单图片查询测试
	fmt.Println("\n📊 Running Single Image Query Test...")
	result = tester.RunSingleImageTest()
	results = append(results, result)
	printTestResult(result)
	
	// 4. 文件上传测试（如果有测试图片）
	if fileExists(testImagePath) {
		fmt.Println("\n📊 Running File Upload Test...")
		result = tester.RunUploadTest()
		results = append(results, result)
		printTestResult(result)
	} else {
		fmt.Printf("⚠️  Skipping upload test - test image not found: %s\n", testImagePath)
	}
	
	// 5. 并发混合负载测试
	fmt.Println("\n📊 Running Mixed Concurrent Load Test...")
	result = tester.RunMixedLoadTest()
	results = append(results, result)
	printTestResult(result)
	
	// 生成测试报告
	fmt.Println("\n📈 Generating Performance Report...")
	generateReport(results)
	
	fmt.Println("\n✅ Performance testing completed!")
}

// 健康检查负载测试
func (pt *PerformanceTester) RunHealthCheckTest() TestResult {
	return pt.runLoadTest("Health Check Load Test", func() error {
		resp, err := pt.client.Get(pt.config.BaseURL + "/health")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		
		return nil
	}, 200, 30*time.Second)
}

// 图片列表查询测试
func (pt *PerformanceTester) RunImageListTest() TestResult {
	return pt.runLoadTest("Image List Query Test", func() error {
		req, err := http.NewRequest("GET", pt.config.BaseURL+"/api/v1/images", nil)
		if err != nil {
			return err
		}
		
		if pt.config.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+pt.config.AuthToken)
		}
		
		resp, err := pt.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		
		return nil
	}, 150, 45*time.Second)
}

// 单图片查询测试
func (pt *PerformanceTester) RunSingleImageTest() TestResult {
	return pt.runLoadTest("Single Image Query Test", func() error {
		// 使用一个假设存在的图片ID或key
		imageKey := "test-image-key"
		req, err := http.NewRequest("GET", pt.config.BaseURL+"/api/v1/images/view/"+imageKey, nil)
		if err != nil {
			return err
		}
		
		resp, err := pt.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		// 接受200或404，因为测试图片可能不存在
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		
		return nil
	}, 200, 30*time.Second)
}

// 文件上传测试
func (pt *PerformanceTester) RunUploadTest() TestResult {
	return pt.runLoadTest("File Upload Test", func() error {
		return pt.uploadTestFile()
	}, 20, 60*time.Second) // 较少的并发数和较长的超时时间
}

// 混合负载测试
func (pt *PerformanceTester) RunMixedLoadTest() TestResult {
	return pt.runLoadTest("Mixed Concurrent Load Test", func() error {
		// 随机选择不同的操作
		switch time.Now().UnixNano() % 3 {
		case 0:
			// 健康检查
			resp, err := pt.client.Get(pt.config.BaseURL + "/health")
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
			
		case 1:
			// 图片列表查询
			req, err := http.NewRequest("GET", pt.config.BaseURL+"/api/v1/images", nil)
			if err != nil {
				return err
			}
			if pt.config.AuthToken != "" {
				req.Header.Set("Authorization", "Bearer "+pt.config.AuthToken)
			}
			resp, err := pt.client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
			
		default:
			// 监控指标查询
			req, err := http.NewRequest("GET", pt.config.BaseURL+"/api/v1/monitoring/metrics", nil)
			if err != nil {
				return err
			}
			if pt.config.AuthToken != "" {
				req.Header.Set("Authorization", "Bearer "+pt.config.AuthToken)
			}
			resp, err := pt.client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
		}
	}, 100, 120*time.Second)
}

// 通用负载测试函数
func (pt *PerformanceTester) runLoadTest(testName string, testFunc func() error, requests int, duration time.Duration) TestResult {
	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	var successCount, failCount int
	var totalLatency, minLatency, maxLatency time.Duration
	minLatency = time.Hour // 设置一个很大的初始值
	
	concurrency := pt.config.Concurrency
	if requests < concurrency {
		concurrency = requests
	}
	
	reqChan := make(chan struct{}, requests)
	for i := 0; i < requests; i++ {
		reqChan <- struct{}{}
	}
	close(reqChan)
	
	// 启动工作协程
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for range reqChan {
				// 检查是否超时
				if time.Since(start) > duration {
					return
				}
				
				reqStart := time.Now()
				err := testFunc()
				latency := time.Since(reqStart)
				
				mu.Lock()
				totalLatency += latency
				if latency < minLatency {
					minLatency = latency
				}
				if latency > maxLatency {
					maxLatency = latency
				}
				
				if err != nil {
					failCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	testDuration := time.Since(start)
	
	totalReqs := successCount + failCount
	var avgLatency time.Duration
	if totalReqs > 0 {
		avgLatency = totalLatency / time.Duration(totalReqs)
	}
	
	var errorRate float64
	if totalReqs > 0 {
		errorRate = float64(failCount) / float64(totalReqs) * 100
	}
	
	reqsPerSec := float64(totalReqs) / testDuration.Seconds()
	
	return TestResult{
		TestName:        testName,
		TotalRequests:   totalReqs,
		SuccessRequests: successCount,
		FailedRequests:  failCount,
		AverageLatency:  avgLatency,
		MinLatency:      minLatency,
		MaxLatency:      maxLatency,
		RequestsPerSec:  reqsPerSec,
		ErrorRate:       errorRate,
		Duration:        testDuration,
	}
}

// 上传测试文件
func (pt *PerformanceTester) uploadTestFile() error {
	file, err := os.Open(pt.config.TestImagePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", "test_image.jpg")
	if err != nil {
		return err
	}
	
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	
	writer.WriteField("is_public", "true")
	writer.WriteField("folder", "test")
	
	err = writer.Close()
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", pt.config.BaseURL+"/api/v1/images/upload", body)
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if pt.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+pt.config.AuthToken)
	}
	
	resp, err := pt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// 打印测试结果
func printTestResult(result TestResult) {
	fmt.Printf("📊 %s Results:\n", result.TestName)
	fmt.Printf("  Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("  Success: %d (%.1f%%)\n", result.SuccessRequests, float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("  Failed: %d (%.1f%%)\n", result.FailedRequests, result.ErrorRate)
	fmt.Printf("  Average Latency: %v\n", result.AverageLatency)
	fmt.Printf("  Min Latency: %v\n", result.MinLatency)
	fmt.Printf("  Max Latency: %v\n", result.MaxLatency)
	fmt.Printf("  Requests/Second: %.2f\n", result.RequestsPerSec)
	fmt.Printf("  Test Duration: %v\n", result.Duration)
	fmt.Println()
}

// 生成性能报告
func generateReport(results []TestResult) {
	reportFile := "performance_report.json"
	
	report := map[string]interface{}{
		"timestamp":        time.Now(),
		"summary":          generateSummary(results),
		"detailed_results": results,
		"recommendations":  generateRecommendations(results),
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error generating report: %v\n", err)
		return
	}
	
	err = os.WriteFile(reportFile, data, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing report file: %v\n", err)
		return
	}
	
	fmt.Printf("📄 Performance report saved to: %s\n", reportFile)
	printSummary(results)
}

// 生成测试摘要
func generateSummary(results []TestResult) map[string]interface{} {
	var totalRequests, totalSuccess, totalFailed int
	var totalLatency time.Duration
	var maxRPS float64
	
	for _, result := range results {
		totalRequests += result.TotalRequests
		totalSuccess += result.SuccessRequests
		totalFailed += result.FailedRequests
		totalLatency += result.AverageLatency
		
		if result.RequestsPerSec > maxRPS {
			maxRPS = result.RequestsPerSec
		}
	}
	
	avgLatency := time.Duration(0)
	if len(results) > 0 {
		avgLatency = totalLatency / time.Duration(len(results))
	}
	
	successRate := float64(0)
	if totalRequests > 0 {
		successRate = float64(totalSuccess) / float64(totalRequests) * 100
	}
	
	return map[string]interface{}{
		"total_requests":           totalRequests,
		"success_requests":         totalSuccess,
		"failed_requests":          totalFailed,
		"success_rate":             successRate,
		"average_latency":          avgLatency.String(),
		"max_requests_per_second":  maxRPS,
		"tests_count":              len(results),
	}
}

// 生成性能建议
func generateRecommendations(results []TestResult) []string {
	var recommendations []string
	
	for _, result := range results {
		if result.ErrorRate > 5 {
			recommendations = append(recommendations, fmt.Sprintf("%s has high error rate (%.1f%%) - investigate error causes", result.TestName, result.ErrorRate))
		}
		
		if result.AverageLatency > 2*time.Second {
			recommendations = append(recommendations, fmt.Sprintf("%s has high latency (%v) - consider optimization", result.TestName, result.AverageLatency))
		}
		
		if result.RequestsPerSec < 10 {
			recommendations = append(recommendations, fmt.Sprintf("%s has low throughput (%.2f RPS) - consider scaling", result.TestName, result.RequestsPerSec))
		}
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Overall performance looks good! 🎉")
	}
	
	return recommendations
}

// 打印摘要
func printSummary(results []TestResult) {
	summary := generateSummary(results)
	recommendations := generateRecommendations(results)
	
	fmt.Println("\n📈 Performance Test Summary:")
	fmt.Printf("  Total Tests: %v\n", summary["tests_count"])
	fmt.Printf("  Total Requests: %v\n", summary["total_requests"])
	fmt.Printf("  Success Rate: %.1f%%\n", summary["success_rate"])
	fmt.Printf("  Average Latency: %v\n", summary["average_latency"])
	fmt.Printf("  Max RPS: %.2f\n", summary["max_requests_per_second"])
	
	fmt.Println("\n💡 Recommendations:")
	for i, rec := range recommendations {
		fmt.Printf("  %d. %s\n", i+1, rec)
	}
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}