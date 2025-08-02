#!/bin/bash
# ci.sh - Script to run CI tests

set -e 

echo "Running citest"

go test -v -timeout 30s -run ^TestAutoCapture$ ./backend/extras/autocdc


go test -v -timeout 30s -run ^TestBinds$ ./backend/engine/luaz/binds