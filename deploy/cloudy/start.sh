#!/bin/bash
set -e

pwd

ls -la

file /app/cloudy

ldd /app/cloudy


echo "Starting cloudy"
exec /app/cloudy
echo "Cloudy exited with code $?"


sleep Infinity
