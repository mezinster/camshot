#!/bin/bash
echo "Building CamShot Puzzle Maker..."

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Build
echo "Compiling..."
go build -o camshot .

if [ $? -eq 0 ]; then
    echo ""
    echo "Build successful! Run ./camshot to start the application."
    echo "Note: Camera capture is only available on Windows."
else
    echo ""
    echo "Build failed. Please check the error messages above."
fi
