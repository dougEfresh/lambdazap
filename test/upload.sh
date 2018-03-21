#!/bin/bash
OP=create
REGION=${1:?}
stackName=${2:?}
awsRun="aws --region $REGION"
bucket="${stackName}-${REGION}"

rm -f handler
go build handler.go
rm -f $stackName.zip 2> /dev/null
zip $stackName.zip  handler

$awsRun s3 mb s3://$bucket 2> /dev/null
$awsRun s3 cp $stackName.zip s3://$bucket/
