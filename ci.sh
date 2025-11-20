#!/bin/bash
# ci.sh - Script to run CI tests

set -e 

echo "Running citest"

RUNINTEGRATIONTESTS=1


if [ "$RUNINTEGRATIONTESTS" -eq 1 ]; then
  echo "Running integration tests"
  go run ./tests/*.go
  go test ./backend/engine/executors/luaz/binds/... -v
else
  echo "Skipping integration tests"
fi