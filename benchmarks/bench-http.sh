#!/bin/bash

set -e

echo ""
echo "--- BENCH HTTP START ---"
echo ""

cd $(dirname "${BASH_SOURCE[0]}")
function cleanup {
    echo "--- BENCH HTTP DONE ---"
    kill -9 $(jobs -rp)
    wait $(jobs -rp) 2>/dev/null
}
trap cleanup EXIT

mkdir -p bin
$(pkill -9 go-http-server || printf "")
$(pkill -9 netti-http-server || printf "")

function gobench {
    echo "--- $1 ---"
    if [[ "$3" != "" ]]; then
        go build -o $2 $3
    fi

    GOMAXPROCS=8 $2 --port $4 &

    sleep 1
    echo "*** 10000 connections, 10 seconds"
    ../tool/bombardier-linux-amd64 -c 10000 -n 1000000 -l http://127.0.0.1:$4
    echo "--- DONE ---"
    echo ""
}

gobench "GO-HTTP" bin/go-http-server-go ../examples/http-server-go/main.go 8081
gobench "NETTI" bin/netti-http-server ../examples/http-server/main.go 8084
