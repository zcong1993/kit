#!/bin/bash
set -e

for example in $(ls ./_examples)
do
  if [[ "$example" != "otel" ]]; then
    echo "build $example"
    go build -o "./bin/${example}" "./_examples/${example}"
  fi
done

for example in $(ls ./_examples/otel)
do
  echo "build $example"
  go build -o "./bin/otel/${example}" "./_examples/otel/${example}"
done
