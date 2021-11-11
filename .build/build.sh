#!/bin/bash

## Utilizes functions
checkCommand() {
  command=$1

  echo "Checking command $command"
  if command -v $command >/dev/null 2>&1; then
    return 1
  fi

  return 0
}

## Pre-check
checkCommand "go"
if [ $? -ne 1 ]; then
  echo "go command isn't detected on your machine, install go firstly"
fi
checkCommand "builder"
if [ $? -ne 1 ]; then
  go install go.opentelemetry.io/collector/cmd/builder@latest
fi

## Build
# check config.yaml exists or not
- mkdir ./buildcache
if [ ! -f "otelcol-builder.yaml" ]; then
  echo "otelcol-builder.yaml doesn't exist, please create it firstly"
  exit 1
fi

# If you build failed, you can try build by yourself in buildcache directory:
# 1. cd path/to/buildcache
# 2. go mod tidy && go build -o otelcol .
#
# automatic execute build command
builder --config=./otelcol-builder.yaml --output-path=./buildcache
cp ./buildcache/otelcol .
#rm -fr ./buildcache
