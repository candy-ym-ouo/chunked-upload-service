#!/bin/sh
files=$(find . -name '*.go' ! -name '*_test.go' | wc -l)
lines=$(find . -name '*.go' ! -name '*_test.go' -print0 | xargs -0 wc -l | tail -1 | awk '{print $1}')
echo "non-test go files: $files, lines: $lines"
