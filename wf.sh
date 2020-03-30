#!/usr/bin/env bash

export PORT=8080

if [ "$1" = "s" ] || [ "$1" = "server" ]; then # ./wf.sh s
    if [ "$2" = "-p" ]; then # ./wf.sh s -p 3333
        if [ -z "$3" ]; then
            echo "ERROR: Port number missing."; exit 1
        elif ! [[ "$3" =~ ^[0-9]+$ ]]; then
            echo "ERROR: Invalid input for port number." exit 1
        else
            PORT="$3"
            go run ./cmd/wf_email_microservice
        fi
    else
        go run ./cmd/wf_email_microservice
    fi
elif [ "$1" = "cproto" ]; then # ./wf.sh cproto example
    if [ -f "./api/proto/v1/$2.proto" ]; then
        protoc -I=. --go_out=./pkg/api/v1 ./api/proto/v1/$2.proto
        echo "SUCCESS: Created file ./pkg/api/v1/$2.pb.go"
    else
        echo "ERROR: Could not find file ./api/proto/v1/$2.proto"; exit 1
    fi
fi 