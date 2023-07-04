# SSH As A service (sshaas)

This is a simple aws lambda go function that runs ssh commands for you. It's good for testing, and for running commands on remote servers. It's usage may be grounds for firing, given that you're passing SSH keys around via http. 

I personally use it from Zapier (since Zapier doesn't have an SSH step) when I don't have time to add rest endpoints to a service, and I just want to run a management command on a remote server.

This service includes basic password authentication in an attempt to avoid basic  ddos. But you can still get ddos'd. Put this behind AWS API Gateway if you actually want to use this in production.

# Build

To build and release (assumes that the function has already deployed to AWS at least once)
```bash
./build.sh function_name your_very_secret_and_long_password optional-aws-region
```

To only build
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go 
chmod +x main
zip $function_name.zip main
```


# Usage
Invoke your lambda from anywhere you have access to AWS.

The sample in `./test/main.go` does this. Usage:

```bash
cd test
go run main.go \
    -ssh_command="echo 'your command'" \
    -ssh_key_file="/home/user/.ssh/key.pem" \
    -lambda_password="your_very_secret_and_long_password" \
    -function_name=function_name \
    -ip_address=ssh.server.ip.address \
    -region=us-east-1 \
    -user=ubuntu
```