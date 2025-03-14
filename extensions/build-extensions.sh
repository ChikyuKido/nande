#!/bin/bash

mkdir -p ../extension-build
for dir in */; do
  if [ -d "$dir" ]; then
    mkdir -p ../extension-build/"$dir"

    cd "$dir" || continue
    go build -o ../../extension-build/"$dir"
    cp .env ../../extension-build/"$dir"
    echo "Built $dir extension"
    cd ..
  fi
done
