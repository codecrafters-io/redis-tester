#!/bin/sh
find "." -type f -name "*.rdb" -exec rm {} +
exec redis-server --loglevel nothing $@
