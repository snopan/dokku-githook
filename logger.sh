log() {
    local url="$1"
    local message="$2"

    curl -X POST -H "Content-Type: application/json" $url -d "{\"content\":\"$message\"}"
}

logCode() {
    local url="$1"
    local message=$(cat)
    local parsedNewLineMessage=${message//$'\n'/\\n}
    local jsonMessage=$(echo $parsedNewLineMessage | sed 's/"/\\"/g')
    echo $jsonMessage
    curl -X POST -H "Content-Type: application/json" $url -d "{\"content\":\"\`\`\`$jsonMessage\`\`\`\"}"
}