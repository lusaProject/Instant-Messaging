#!/bin/bash

SERVICE_NAME="demo"

check_service() {
    
    if ! pgrep -x "$SERVICE_NAME" > /dev/null; then
        nohup /root/demo/demo &
        echo "$SERVICE_NAME restarted at $(date)"
    fi
}

while true; do
    check_service
    sleep 3
done

