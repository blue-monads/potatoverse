#!/bin/bash

# Demo reset script for potatoverse
# - Deletes .pdata/maindb directory (database state, keeps config intact)
# - Runs server in a loop
# - Resets demo every 2 hours
# - Handles crashes and restarts automatically
#
# Usage:
#   ./run_demo.sh                    # Uses default 'potatoverse' binary
#   POTATOVERSE_BIN=./potato ./run_demo.sh  # Uses custom binary

set -e

# Binary to use (can be overridden with POTATOVERSE_BIN environment variable)
POTATOVERSE_BIN="${POTATOVERSE_BIN:-potatoverse}"

RESET_INTERVAL=7200  # 2 hours in seconds
PID_FILE="/tmp/potatoverse_demo.pid"
LOG_FILE="/tmp/potatoverse_demo.log"
MONITOR_PID=""
TIMER_PID=""

# Function to cleanup on exit
cleanup() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaning up..." | tee -a "$LOG_FILE"
    
    # Stop monitor and timer processes
    if [ -n "$MONITOR_PID" ] && kill -0 "$MONITOR_PID" 2>/dev/null; then
        kill "$MONITOR_PID" 2>/dev/null || true
    fi
    if [ -n "$TIMER_PID" ] && kill -0 "$TIMER_PID" 2>/dev/null; then
        kill "$TIMER_PID" 2>/dev/null || true
    fi
    
    # Stop server
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            # Kill the process and all its children
            pkill -P "$PID" 2>/dev/null || true
            kill "$PID" 2>/dev/null || true
            wait "$PID" 2>/dev/null || true
        fi
        rm -f "$PID_FILE"
    fi
    
    exit 0
}

# Trap signals for cleanup
trap cleanup SIGINT SIGTERM EXIT

# Function to delete .pdata/maindb directory (keeps config intact)
reset_state() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Resetting demo state (deleting .pdata/maindb)..." | tee -a "$LOG_FILE"
    if [ -d ".pdata/maindb" ]; then
        rm -rf ".pdata/maindb"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] .pdata/maindb directory deleted" | tee -a "$LOG_FILE"
    else
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] .pdata/maindb directory not found (already clean)" | tee -a "$LOG_FILE"
    fi
}

# Function to stop server if running
stop_server() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] Stopping server (PID: $PID)..." | tee -a "$LOG_FILE"
            # Kill the process and all its children
            pkill -P "$PID" 2>/dev/null || true
            kill "$PID" 2>/dev/null || true
            wait "$PID" 2>/dev/null || true
        fi
        rm -f "$PID_FILE"
    fi
}

# Function to find the actual server process (child process listening on port 7777)
find_server_pid() {
    # Wait a moment for the server to start
    sleep 3
    
    # Try to find process listening on port 7777
    local pid=$(lsof -ti:7777 2>/dev/null || true)
    
    if [ -n "$pid" ]; then
        echo "$pid"
        return 0
    fi
    
    # Fallback: try to find potatoverse process with "actual-start" in command line
    pid=$(pgrep -f "$POTATOVERSE_BIN server actual-start" | head -1 || true)
    
    if [ -n "$pid" ]; then
        echo "$pid"
        return 0
    fi
    
    return 1
}

# Function to start server
start_server() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Initializing server..." | tee -a "$LOG_FILE"
    
    # Initialize server
    if ! "$POTATOVERSE_BIN" server init >> "$LOG_FILE" 2>&1; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: Server init failed" | tee -a "$LOG_FILE"
        return 1
    fi
    
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting server..." | tee -a "$LOG_FILE"
    
    # Start server in background
    "$POTATOVERSE_BIN" server start >> "$LOG_FILE" 2>&1 &
    PARENT_PID=$!
    
    # Find the actual server process (child process)
    SERVER_PID=$(find_server_pid)
    
    if [ -z "$SERVER_PID" ] || [ "$SERVER_PID" = "" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: Could not find server process after start" | tee -a "$LOG_FILE"
        # Check if parent process is still running
        if kill -0 "$PARENT_PID" 2>/dev/null; then
            kill "$PARENT_PID" 2>/dev/null || true
        fi
        return 1
    fi
    
    echo "$SERVER_PID" > "$PID_FILE"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Server started with PID: $SERVER_PID (parent: $PARENT_PID)" | tee -a "$LOG_FILE"
    
    # Verify the process is still running
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: Server crashed immediately after start" | tee -a "$LOG_FILE"
        rm -f "$PID_FILE"
        return 1
    fi
    
    return 0
}

# Function to monitor server and restart on crash
monitor_server() {
    while true; do
        if [ -f "$PID_FILE" ]; then
            PID=$(cat "$PID_FILE")
            if ! kill -0 "$PID" 2>/dev/null; then
                echo "[$(date '+%Y-%m-%d %H:%M:%S')] Server process died (PID: $PID), restarting..." | tee -a "$LOG_FILE"
                rm -f "$PID_FILE"
                sleep 2
                if ! start_server; then
                    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Failed to restart server, retrying in 5 seconds..." | tee -a "$LOG_FILE"
                    sleep 5
                fi
            fi
        else
            # No PID file, try to start server
            if ! start_server; then
                echo "[$(date '+%Y-%m-%d %H:%M:%S')] Failed to start server, retrying in 5 seconds..." | tee -a "$LOG_FILE"
                sleep 5
            fi
        fi
        sleep 5  # Check every 5 seconds
    done
}

# Function to handle periodic resets
reset_timer() {
    while true; do
        sleep "$RESET_INTERVAL"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] 2-hour timer expired, resetting demo..." | tee -a "$LOG_FILE"
        stop_server
        reset_state
        sleep 2
        start_server
    done
}

# Main execution
echo "=========================================" | tee -a "$LOG_FILE"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting potatoverse demo script" | tee -a "$LOG_FILE"
echo "=========================================" | tee -a "$LOG_FILE"

# Initial reset
reset_state

# Start server
if ! start_server; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Failed to start server initially, exiting" | tee -a "$LOG_FILE"
    exit 1
fi

# Start monitor in background
monitor_server &
MONITOR_PID=$!

# Start reset timer in background
reset_timer &
TIMER_PID=$!

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Demo script running. Monitor PID: $MONITOR_PID, Timer PID: $TIMER_PID" | tee -a "$LOG_FILE"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Server will reset every 2 hours" | tee -a "$LOG_FILE"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Log file: $LOG_FILE" | tee -a "$LOG_FILE"
echo "Press Ctrl+C to stop" | tee -a "$LOG_FILE"

# Wait for background processes (they run indefinitely)
wait $MONITOR_PID $TIMER_PID 2>/dev/null || true
