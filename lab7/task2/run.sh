#!/bin/sh
# Participation task: who is printed first? 10 runs with the default GOMAXPROCS,
# then 10 runs with GOMAXPROCS=1. One line per run.
# Usage: ./task2/run.sh   (from the lab7 directory)
set -e
bin=$(mktemp -d)/order
go build -o "$bin" ./task2
echo "GOMAXPROCS=default:"
for i in 1 2 3 4 5 6 7 8 9 10; do "$bin" order | grep -o '[0-9]' | tr -d '\n'; echo; done
echo "GOMAXPROCS=1:"
for i in 1 2 3 4 5 6 7 8 9 10; do GOMAXPROCS=1 "$bin" order | grep -o '[0-9]' | tr -d '\n'; echo; done
rm -r "$(dirname "$bin")"
