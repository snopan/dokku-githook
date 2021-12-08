#!/bin/bash

start() {
	declare GITHUB_USERNAME=$1
	declare GITHUB_TOKEN=$2
	declare GITHUB_HOOK_PORT=$3
	declare LOCAL_CONTROL_PORT=$4
	echo "$GITHUB_USERNAME $GITHUB_TOKEN $GITHUB_HOOK_PORT $LOCAL_CONTROL_PORT"
	/home/dokku/go/bin/go run ./main.go	
}

start "$@"