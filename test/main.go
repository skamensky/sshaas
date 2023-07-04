package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// sample usage:
// go run main.go -ssh_command="ls -a" -ssh_key_file="/home/user/.ssh/key.pem" -lambda_password="your_very_secret_and_long_password" -function_name=your_lambda_function_name -ip_address=ssh.server.ip.address -user=ssh_user -region=optional-region

func main() {
	var sshCommand, sshKeyFile, lambda_password, functionName, ip_address, region, user string

	flag.StringVar(&sshCommand, "ssh_command", "", "SSH command to be executed")
	flag.StringVar(&sshKeyFile, "ssh_key_file", "", "Path to the SSH key file")
	flag.StringVar(&functionName, "function_name", "", "Function name")
	flag.StringVar(&lambda_password, "lambda_password", "", "Lambda password")
	flag.StringVar(&ip_address, "ip_address", "", "IP address")
	flag.StringVar(&user, "user", "", "SSH user")
	flag.StringVar(&region, "region", "us-east-1", "aws region")
	flag.Parse()

	if sshCommand == "" {
		fmt.Println("Missing SSH command")
		os.Exit(1)
	}
	if sshKeyFile == "" {
		fmt.Println("Missing SSH key file")
		os.Exit(1)
	}
	if functionName == "" {
		fmt.Println("Missing function name")
		os.Exit(1)
	}
	if lambda_password == "" {
		fmt.Println("Missing Lambda password")
		os.Exit(1)
	}
	if ip_address == "" {
		fmt.Println("Missing IP address")
		os.Exit(1)
	}
	if user == "" {
		fmt.Println("Missing SSH user")
		os.Exit(1)
	}

	// Read SSH key from file
	sshKey, err := ioutil.ReadFile(sshKeyFile)
	if err != nil {
		fmt.Println("Error reading SSH key file:", err)
		os.Exit(1)
	}

	// Construct the payload
	payload, err := json.Marshal(map[string]string{
		"command":         sshCommand,
		"user":            user,
		"ip_address":      ip_address,
		"ssh_key":         string(sshKey),
		"lambda_password": lambda_password,
	})
	if err != nil {
		fmt.Println("Error constructing payload:", err)
		os.Exit(1)
	}

	// Create a Lambda client
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		os.Exit(1)
	}
	lambdaClient := lambda.New(sess)

	// Invoke the Lambda function
	response, err := lambdaClient.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	})
	if err != nil {
		fmt.Println("Error invoking Lambda function:", err)
		os.Exit(1)
	}

	// Write the response to a file
	err = ioutil.WriteFile("response.json", response.Payload, 0644)
	if err != nil {
		fmt.Println("Error writing response to file:", err)
		os.Exit(1)
	}

	// Read the response from the file and print the result
	responseData, err := ioutil.ReadFile("response.json")
	if err != nil {
		fmt.Println("Error reading response from file:", err)
		os.Exit(1)
	}
	var responsePayload map[string]interface{}
	err = json.Unmarshal(responseData, &responsePayload)
	if err != nil {
		fmt.Println("Error unmarshalling response data:", err)
		os.Exit(1)
	}
	fmt.Println(responsePayload["result"])
}
