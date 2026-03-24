#!/bin/bash

# Build Script
# Compiles the Go CLI application

echo "Building CLI application..."

go build -o order-controller ./main.go

echo "Build completed"