package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:  "run-go-lambda",
		RunE: invoke,
	}
	timeout      int64
	payloadFile  string
	payloadStdIn []byte
)

// initialize command options
func init() {
	rootCmd.Flags().Int64VarP(&timeout, "timeout", "t", 300, "duration of timeout")
	rootCmd.Flags().StringVarP(&payloadFile, "file", "f", "", "JSON file")
}

func readStdIn() []byte {
	file := os.Stdin
	fi, _ := file.Stat()
	size := fi.Size()
	if size != 0 {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		return data
	}
	return nil
}

// invoke the lambda
func invoke(cmd *cobra.Command, args []string) error {
	data, err := readPayload()
	if err != nil {
		return err
	}
	req := &messages.InvokeRequest{
		Payload:            data,
		RequestId:          "1",
		XAmznTraceId:       "1",
		Deadline:           messages.InvokeRequest_Timestamp{Seconds: timeout, Nanos: 0},
		InvokedFunctionArn: "arn:aws:lambda:an-antarctica-1:123456789100:function:test",
	}

	client := connect()

	var response *messages.InvokeResponse
	err = client.Call("Function.Invoke", req, &response)

	if err != nil {
		log.Println("Invocation:", err)
		log.Fatal("Response:", response)
		return err
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}

func readPayload() ([]byte, error) {
	payloadStdIn = readStdIn()
	if len(payloadStdIn) != 0 {
		return payloadStdIn, nil
	}
	if payloadFile != "" {
		payload, err := ioutil.ReadFile(payloadFile)
		if err != nil {
			log.Fatal(err)
		}
		return payload, nil
	}
	return nil, errors.New("requires at least a file input a stdin JSON")
}

func connect() *rpc.Client {
	port := os.Getenv("_LAMBDA_SERVER_PORT")
	serverAddress := fmt.Sprintf("localhost:%s", port)
	log.Println("Test harness connecting to: " + serverAddress)

	var client *rpc.Client
	connect := func() error {
		var err error
		client, err = rpc.Dial("tcp", serverAddress)
		if err != nil {
			return err
		}
		return nil
	}
	err := backoff.Retry(connect, constantBackoff())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func constantBackoff() *backoff.ExponentialBackOff {
	algorithm := backoff.NewExponentialBackOff()
	algorithm.MaxElapsedTime = 8 * time.Second
	algorithm.Multiplier = 1
	return algorithm
}
