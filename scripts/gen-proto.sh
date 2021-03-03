#!/bin/bash -x
set -e

OUT_DIR=./pb

rm -rf $OUT_DIR/*.{go,json}

protoc --go-grpc_out=. \
    --go_out=. \
    -I=. \
    ./_examples/grpc/pb/*.proto
