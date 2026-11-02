#!/bin/sh
# Runs the GC demo with different settings, each 3 times.
# Usage: ./task6/run.sh   (from the lab6 directory)
set -e
go build -o /tmp/gcdemo ./task6
for g in 50 100 200 400 off; do
	for i in 1 2 3; do GOGC=$g /tmp/gcdemo; done
done
for i in 1 2 3; do GOGC=off GOMEMLIMIT=100MiB /tmp/gcdemo; done
