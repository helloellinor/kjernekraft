#!/bin/bash

# Shuffle Test Data Script
# This script regenerates randomized test data for the Kjernekraft application

echo "🔄 Shuffling test data for Kjernekraft dashboard..."
echo "=================================================="

# Navigate to the test directory and run the test data manager
cd "$(dirname "$0")"
go run test_data_manager.go

echo ""
echo "✨ Done! You can now:"
echo "   1. Start the server: go run ../server.go"
echo "   2. Visit http://localhost:8080/elev/hjem to see today's classes"
echo "   3. Visit http://localhost:8080/elev/timeplan to see this week's schedule"
echo "   4. Run this script again anytime to get new randomized data!"