#!/bin/bash

start() {
	export GITHUB_USERNAME=$1
	export GITHUB_TOKEN=$2
	export GITHUB_HOOK_PORT=$3
	export LOCAL_CONTROL_PORT=$4
	echo "testing $GITHUB_USERNAME"
	/home/dokku/go/bin/go run ./main.go	$GITHUB_USERNAME $GITHUB_TOKEN
}

start "$@"