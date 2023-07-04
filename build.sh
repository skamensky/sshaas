#! /bin/bash

# this script assumes that the function already exists with a role already assigned


set -e


# used https://stackoverflow.com/a/34676160/4188138 for temporary directory

USAGE="Usage: $0 function_name lambda_password [region]
The default region is us-east-1
"


function_name=$1
lambda_password=$2
region=$3
if [ -z "$function_name" ]; then
    echo "Function name is required. $USAGE"
    exit 1
fi

if [ -z "$lambda_password" ]; then
    echo "lambda_password is required. $USAGE"
    exit 1
fi

if [ -z "$region" ]; then
    region="us-east-1"
    echo "Region not specified. Using default region $region"
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
WORK_DIR=`mktemp -d -p "$DIR"`
JQ_EXISTS=`which jq`
function cleanup {      
  rm -rf "$WORK_DIR"
  cd "$DIR"
}

trap cleanup EXIT

echo "Building and zipping main.go"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go 
chmod +x main
cd "$WORK_DIR"
mv ../main .
zip $function_name.zip main

echo "uploading $function_name.zip to function $function_name in region $region "
if [ -z "$JQ_EXISTS" ]; then
    aws --region $region lambda update-function-code --function-name $function_name --zip-file fileb://$function_name.zip
else
    aws --region $region lambda update-function-code --function-name $function_name --zip-file fileb://$function_name.zip |jq
fi
echo "uploaded $function_name.zip. Waiting for function to be updated..."
aws lambda wait function-exists --function-name $function_name --region $region
echo "function $function_name created/updated successfully. Adding lambda_password to environment variables"
has_vars=$(aws lambda get-function-configuration --region $region  --function-name $function_name --query "Environment.Variables")
if [ "$has_vars" == "null" ]; then
    NEW_ENVVARS="{\"LAMBDA_PASSWORD\":\"$lambda_password\"}"
else
    # from https://stackoverflow.com/a/71157145/4188138 
    echo "Environment variables found. Merging lambda_password with current environment variables"
    NEW_ENVVARS=$(aws lambda get-function-configuration --region $region  --function-name $function_name --query "Environment.Variables | merge(@, \`{\"LAMBDA_PASSWORD\":\"$lambda_password\"}\`)")
    
fi
echo "New environment variables: $NEW_ENVVARS"
if [ -z "$JQ_EXISTS" ]; then
    aws lambda update-function-configuration --region $region --function-name $function_name --environment "{ \"Variables\": $NEW_ENVVARS }"
else
    aws lambda update-function-configuration --region $region --function-name $function_name --environment "{ \"Variables\": $NEW_ENVVARS }" | jq
fi
