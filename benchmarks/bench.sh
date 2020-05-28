#!/bin/bash

set -e

cd $(dirname "${BASH_SOURCE[0]}")

mkdir -p results/

./bench-http.sh 2>&1 | tee results/http.txt

go run analyze.go