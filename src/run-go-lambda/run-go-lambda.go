package main

import (
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
		RunE: run,
	}
	timeout      int64
	payloadFile  string
)

// initialize command options
func init() {
	rootCmd.Flags().Int64VarP(&timeout, "timeout", "t", 300, "duration of timeout")
	rootCmd.Flags().StringVarP(&payloadFile, "file", "f", "", "JSON file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	var data []byte
	var err error

	if cmd.Flag("file").Changed {
		if payloadFile == "" {
			fmt.Fprintln(os.Stderr, "Error: You must specify a filename if you pass the -f/--file flag")
			os.Exit(1)
		}
		data, err = readInputFile()
	} else {
		data = readStdIn()
	}
	if err != nil {
		return err
	}

	return invoke(data)
}

func readInputFile() ([]byte, error) {
	payload, err := ioutil.ReadFile(payloadFile)
	if err != nil {
		log.Fatal(err)
	}
	return payload, nil
}

func readStdIn() []byte {
	file := os.Stdin
	fi, _ := file.Stat()
	size := fi.Size()
	if size == 0 {
		log.Fatal("Error: Expecting input on stdin, if no -f/--file flag passed.")
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func invoke(data []byte) error {
	req := &messages.InvokeRequest{
		Payload:            data,
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
	}
	return nil
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
