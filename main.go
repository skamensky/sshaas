package main

import (
	"context"
	"encoding/json"
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
	requestAsJson, err := json.Marshal(request)
	if err == nil {
		fmt.Println("Request: ", string(requestAsJson))
	}
	errMessage := ""

	lambda_server_password := os.Getenv("LAMBDA_PASSWORD")

	if lambda_server_password == "" {
		errMessage = "server is missing lambda_password"
		fmt.Println(errMessage)
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}

	if request.LambdaPassword == "" {
		errMessage = "missing lambda_password"
		fmt.Println(errMessage)
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}
	if request.LambdaPassword != lambda_server_password {
		// don't waste time on further checks if the password is incorrect
		errMessage = "incorrect lambda_password"
		fmt.Println(errMessage)
		return Response{
			StatusCode: 403,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}
	if request.IPAddress == "" {
		errMessage = "missing ip_address"
		fmt.Println(errMessage)
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}
	if request.SSHKey == "" {
		errMessage = "missing ssh_key"
		fmt.Println(errMessage)
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}

	if request.Command == "" {
		errMessage = "missing command"
		fmt.Println(errMessage)
		return Response{
			StatusCode: 400,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}

	keyBytes := []byte(request.SSHKey)
	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		errMessage = fmt.Sprintf("failed to parse private key: %v", err)
		fmt.Println(errMessage)
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf(errMessage)
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
		errMessage = fmt.Sprintf("failed to connect to SSH server: %v", err)
		fmt.Println(errMessage)
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		errMessage = fmt.Sprintf("failed to create SSH session: %v", err)
		fmt.Println(errMessage)
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}
	defer session.Close()

	output, err := session.CombinedOutput(request.Command)
	if err != nil {
		errMessage = fmt.Sprintf("failed to run SSH command: output: %v err: %v", string(output), err)
		fmt.Println(errMessage)
		return Response{
			StatusCode: 200,
			Result:     "",
		}, fmt.Errorf(errMessage)
	}

	fmt.Println(string(output))

	return Response{
		StatusCode: 200,
		Result:     string(output),
	}, nil

}

func main() {
	lambda.Start(HandleRequest)
}
