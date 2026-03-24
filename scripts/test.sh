#!/bin/bash

# Unit Test Script
# Runs all Go unit tests with verbose output

echo "Running unit tests..."

go test ./... -v

echo "Unit tests completed"
