#!/bin/bash
set -euo pipefail

mkdir -p third_party/google/api

if [ ! -f third_party/google/api/annotations.proto ]; then
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > third_party/google/api/annotations.proto
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > third_party/google/api/http.proto
fi

# Generate Go bindings
protoc -I api/proto \
       -I third_party \
       --go_out=. --go_opt=module=github.com/Rioverde/agent-corp \
       --go-grpc_out=. --go-grpc_opt=module=github.com/Rioverde/agent-corp \
       api/proto/auth/auth.proto

# Generate descriptor set for Envoy gRPC-JSON transcoder.
# --include_imports bundles google.api.http annotations Envoy needs.
protoc -I api/proto \
       -I third_party \
       --include_imports \
       --include_source_info \
       --descriptor_set_out=deploy/envoy/proto/auth.pb \
       api/proto/auth/auth.proto
