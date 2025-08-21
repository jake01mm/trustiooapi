#!/bin/bash

# Trusioo API Performance Testing Script
# This script runs comprehensive performance tests against the Trusioo API

set -e

echo "🚀 Trusioo API Performance Testing Suite"
echo "========================================"

# Default configuration
API_BASE_URL=${API_BASE_URL:-"http://localhost:8080"}
AUTH_TOKEN=${AUTH_TOKEN:-""}
TEST_IMAGE_PATH=${TEST_IMAGE_PATH:-"test_image.jpg"}
CONCURRENCY=${CONCURRENCY:-10}
OUTPUT_DIR=${OUTPUT_DIR:-"performance_results"}

echo "📋 Test Configuration:"
echo "  API Base URL: $API_BASE_URL"
echo "  Concurrency: $CONCURRENCY"
echo "  Auth Token: $(if [ -n "$AUTH_TOKEN" ]; then echo "***PROVIDED***"; else echo "NOT SET"; fi)"
echo "  Test Image: $TEST_IMAGE_PATH"
echo "  Output Directory: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if API is accessible
echo "🔍 Checking API accessibility..."
if curl -s --max-time 10 "$API_BASE_URL/health" > /dev/null; then
    echo "✅ API is accessible at $API_BASE_URL"
else
    echo "❌ API is not accessible at $API_BASE_URL"
    echo "Please ensure the API server is running and accessible."
    exit 1
fi

# Create a simple test image if it doesn't exist
if [ ! -f "$TEST_IMAGE_PATH" ]; then
    echo "📷 Creating test image..."
    # Create a simple 1x1 pixel PNG for testing
    echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" | base64 -d > "$TEST_IMAGE_PATH"
    echo "✅ Test image created: $TEST_IMAGE_PATH"
fi

# Set environment variables for the test
export API_BASE_URL
export AUTH_TOKEN  
export TEST_IMAGE_PATH

echo "🏃 Running performance tests..."
echo ""

# Run the performance test
cd "$(dirname "$0")"

# Compile and run the performance test
go build -o performance_test performance_test.go

if ./performance_test; then
    echo ""
    echo "✅ Performance tests completed successfully!"
    
    # Move results to output directory
    if [ -f "performance_report.json" ]; then
        mv performance_report.json "$OUTPUT_DIR/performance_report_$(date +%Y%m%d_%H%M%S).json"
        echo "📊 Performance report saved to: $OUTPUT_DIR/"
    fi
    
    echo ""
    echo "📈 Performance Testing Summary:"
    echo "  - All tests completed"
    echo "  - Results saved to: $OUTPUT_DIR/"
    echo "  - Check the JSON report for detailed metrics"
    
    # Clean up
    rm -f performance_test
    
else
    echo ""
    echo "❌ Performance tests failed!"
    echo "Check the output above for error details."
    rm -f performance_test
    exit 1
fi

echo ""
echo "🎉 Performance testing suite completed!"
echo "Next steps:"
echo "  1. Review the performance report in $OUTPUT_DIR/"
echo "  2. Analyze bottlenecks and optimization opportunities"
echo "  3. Implement recommended improvements"
echo "  4. Re-run tests to verify improvements"