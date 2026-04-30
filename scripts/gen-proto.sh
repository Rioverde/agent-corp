#!/bin/bash

# Create the necessary directories
mkdir -p third_party/google/api

# Download the proto files if they don't exist
if [ ! -f third_party/google/api/annotations.proto ]; then
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > third_party/google/api/annotations.proto
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > third_party/google/api/http.proto
fi

# Generate the Go code
protoc -I api/proto \
       -I third_party \
       --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/auth/auth.proto