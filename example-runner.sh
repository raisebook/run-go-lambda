#!/usr/bin/env bash
# Summary: Run the lambda in the current directory lambda locally, feeding in test.json
#
# Runs the lambda function itself in the background, so it will listen on the RPC PORT
# Then this program will run, try and connect to the RPC port, and feed in the data from the
# payload file provided on the command line (defaults to test.json)

shopt -s extglob

TEST_FILE=${1:-test.json}

RPC_PORT=10101

NODE_ENV=development \
_LAMBDA_SERVER_PORT=${RPC_PORT} \
AWS_LAMBDA_FUNCTION_NAME=$(basename `pwd`) \
AWS_LAMBDA_FUNCTION_VERSION=1 \
go run !(*_test).go &

_LAMBDA_SERVER_PORT=${RPC_PORT} bin/run-go-lambda ${TEST_FILE}

kill %1 # kill the lambda function server
