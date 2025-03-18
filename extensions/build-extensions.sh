#!/bin/sh

mkdir -p ../extension-build
for dir in */; do
  if [ -d "$dir" ]; then
    mkdir -p ../extension-build/"$dir"

    cd "$dir" || continue
    go build -o ../../extension-build/"$dir/run"
    while read -r line; do
        echo "moving $line to build dir"
        cp "$line" ../../extension-build/"$dir"
    done < <(cat .tomove)
    echo "Built $dir extension"
    cd ..
  fi
done
