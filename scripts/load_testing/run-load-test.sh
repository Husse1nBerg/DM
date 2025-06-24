#!/bin/bash

# Load Test Runner with Error Log Extraction
# This script runs the k6 load test and automatically saves the error log

echo "🚀 Starting k6 load test..."

# Create logs directory if it doesn't exist
mkdir -p logs

# Run k6 test and capture all output
echo "📊 Running load test..."
k6 run k6-load-test.js >logs/test-results-$(date +%Y%m%d_%H%M%S).log 2>&1

# Extract error log from the results
echo "📋 Extracting error log..."
latest_log=$(ls -t logs/test-results-*.log | head -1)

if grep -q "ERROR_LOG_START" "$latest_log"; then
    # Extract JSON error log
    sed -n '/ERROR_LOG_START/,/ERROR_LOG_END/p' "$latest_log" | sed '1d;$d' >logs/error-log-$(date +%Y%m%d_%H%M%S).json
    echo "✅ Error log saved to logs/error-log-$(date +%Y%m%d_%H%M%S).json"
else
    echo "🎉 No errors found in the test results!"
fi

echo "📁 Full test results saved to: $latest_log"
echo "🏁 Load test completed!"

# Show summary
echo -e "\n📈 Test Summary:"
echo "=================="
if [ -f "$latest_log" ]; then
    grep -E "(✓|✗|ERROR LOG SUMMARY|Total errors)" "$latest_log" | tail -20
fi
