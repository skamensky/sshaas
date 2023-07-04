package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/crypto/ssh"
)

type RequestEvent struct {
	IPAddress      string `json:"ip_address"`
	User           string `json:"user"`
	LambdaPassword string `json:"lambda_password"`
	SSHKey         string `json:"ssh_key"`
	Command        string `json:"command"`
}

type Response struct {
	StatusCode int    `json:"statusCode"`
	Result     string `json:"result"`
}

func HandleRequest(ctx context.Context, request RequestEvent) (Response, error) {
	lambda_password := os.Getenv("LAMBDA_PASSWORD")
	if lambda_password == "" {
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf("missing lambda_password")
	}
	if request.IPAddress == "" {
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf("missing ip_address")
	}
	if request.SSHKey == "" {
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf("missing ssh_key")
	}

	if request.Command == "" {
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf("missing command")
	}

	if request.LambdaPassword != lambda_password {
		return Response{
			StatusCode: 403,
			Result:     "",
		}, fmt.Errorf("incorrect password")
	}
	keyBytes := []byte(request.SSHKey)
	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf("failed to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            request.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	client, err := ssh.Dial("tcp", request.IPAddress, config)
	if err != nil {
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf("failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(request.Command)
	if err != nil {
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf("failed to run SSH command: output: %v err: %v", string(output), err)
	}

	return Response{
		StatusCode: 200,
		Result:     string(output),
	}, nil

}

func main() {
	lambda.Start(HandleRequest)
}
