#!/bin/bash
# ci.sh - Script to run CI tests

set -e 

echo "Running citest"

RUNUNITTESTS=0
RUNINTEGRATIONTESTS=1


if [ "$RUNUNITTESTS" -eq 1 ]; then
  echo "Running unit tests"
  go test -v -timeout 30s -run ^TestAutoCapture$ ./backend/extras/autocdc
  go test -v -timeout 30s -run ^TestBinds$ ./backend/engine/luaz/binds
else
  echo "Skipping unit tests"
fi

if [ "$RUNINTEGRATIONTESTS" -eq 1 ]; then
  echo "Running integration tests"
  go run ./tests/*.go
else
  echo "Skipping integration tests"
fi