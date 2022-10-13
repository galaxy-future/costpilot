#!/bin/bash
RUN_NAME="costpilot"
mkdir -p output/conf output/bin

find conf/ -type f ! -name "*local*" -print0 | xargs -0 -I{} cp {} output/conf/
cp -rf website/ output/
cp scripts/run.sh output/

go fmt ./...
go vet ./...

CGO_ENABLED=0 GO111MODULE=on go build -o output/bin/${RUN_NAME} ./