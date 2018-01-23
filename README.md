# run-go-lambda

This is a simple tool that can be used to send test data to a Go program written to run with the [AWS Lambda Go Runtime](https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/) [(Announcement)](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/)

When you run (via a binary produced by `go build` or via `go run`) your function, the Lambda handler starts a `net/rpc` server.
This tool can then connect to that server, and invoke the function with a given payload.

This is quite handy for quickly testing your function works (possibly against local resources) before deploying it to AWS.

## Usage

See the [example shell script](example-runner.sh) for a possible approach to using the tool.  

## Building

Do a `dep ensure`and then `go build`. Copy the binary somewhere useful - `go install` might be applicable for for workspace.


## Contributing

Please read the [Licence](LICENCE) and [Code of Conduct](COC.md) before contributing.

Patches can be submitted by following the usual GitHub "fork -> feature branch -> pull request" dance.

Thanks.

