#!/bin/bash

# Stop existing service if running
if pgrep -f "app"; then
    pkill -f "app"
fi

# Start the Go API
nohup ./app > app.log 2>&1 &

echo "API restarted successfully"