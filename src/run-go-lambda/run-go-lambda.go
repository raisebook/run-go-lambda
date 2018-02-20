package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
		Args: func(cmd *cobra.Command, args []string) error {
			readStdIn()
			file := cmd.Flag("file")
			if file.Value.String() == "" && len(payloadStdIn) == 0 {
				return errors.New("requires at least a file input a stdin JSON")
			}
			return nil
		},
	}
	timeout      int64
	payloadFile  string
	payloadStdIn string
)

// initialize command options
func init() {
	rootCmd.Flags().Int64VarP(&timeout, "timeout", "t", 300, "duration of timeout")
	rootCmd.Flags().StringVarP(&payloadFile, "file", "f", "", "JSON file")
}

func readStdIn() {
	file := os.Stdin
	fi, _ := file.Stat()
	reader := bufio.NewReader(os.Stdin)
	size := fi.Size()
	if size != 0 {
		for {
			input, err := reader.ReadString('\n')
			payloadStdIn = payloadStdIn + input
			if err != nil && err == io.EOF {
				break
			}
		}
	}
}

// invoke the lambda
func invoke(cmd *cobra.Command, args []string) error {

	req := &messages.InvokeRequest{
		Payload:            readPayload(),
		RequestId:          "1",
		XAmznTraceId:       "1",
		Deadline:           messages.InvokeRequest_Timestamp{Seconds: timeout, Nanos: 0},
		InvokedFunctionArn: "arn:aws:lambda:an-antarctica-1:123456789100:function:test",
	}

	client := connect()

	var response *messages.InvokeResponse
	err := client.Call("Function.Invoke", req, &response)

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

func readPayload() []byte {
	if payloadFile != "" {
		payload, err := ioutil.ReadFile(payloadFile)
		if err != nil {
			log.Fatal(err)
		}
		return payload
	}
	return []byte(payloadStdIn)
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
