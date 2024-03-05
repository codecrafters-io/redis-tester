#!/bin/sh
echo "Starting Redis server"
find "." -type f -name "*.rdb" -exec rm {} +
exec redis-server $@
