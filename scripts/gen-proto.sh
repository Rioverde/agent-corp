#!/bin/bash
set -euo pipefail

mkdir -p third_party/google/api

if [ ! -f third_party/google/api/annotations.proto ]; then
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > third_party/google/api/annotations.proto
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > third_party/google/api/http.proto
fi

protoc -I api/proto \
       -I third_party \
       --go_out=. --go_opt=module=github.com/Rioverde/agent-corp \
       --go-grpc_out=. --go-grpc_opt=module=github.com/Rioverde/agent-corp \
       api/proto/auth/auth.proto
