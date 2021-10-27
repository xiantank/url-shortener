#!/bin/bash

name='redis-redisbloom'

docker start $name 1>/dev/null || docker run -d -p 6379:6379 --name $name redislabs/rebloom:latest
