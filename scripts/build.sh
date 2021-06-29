#!/bin/bash
set -e

cd _examples

for example in $(ls .)
do
  if [[ ! "$example" =~ go.(mod|sum)$ ]]; then
    echo "build $example"
    go build -o "../bin/${example}" "./${example}"
  fi
done

cd ../

#for example in $(ls ./_examples/otel)
#do
#  echo "build $example"
#  go build -o "./bin/otel/${example}" "./_examples/otel/${example}"
#done
