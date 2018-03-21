#!/bin/bash -x

OP=create
REGION=eu-central-1
stackName=zap-test
awsRun="aws --region $REGION"
bucket="${stackName}-${REGION}"

./upload.sh $REGION $stackName && \
$awsRun cloudformation ${OP}-stack --stack-name $stackName --template-body file://test-template.yaml --capabilities CAPABILITY_IAM && \
$awsRun cloudformation wait stack-${OP}-complete  --stack-name $stackName && \
$awsRun cloudformation describe-stacks --stack-name $stackName  --query 'Stacks[*].Outputs' && \
./invoke.sh $REGION $stackName
