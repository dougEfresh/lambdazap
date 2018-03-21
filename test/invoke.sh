#!/bin/bash

REGION=${1:?}
stackName=${2:?}
awsRun="aws --region $REGION"
bucket="${stackName}-${REGION}"
output=`mktemp`

$awsRun lambda update-function-code --publish --function-name $stackName --s3-bucket $bucket --query Version  --s3-key ${stackName}.zip
$awsRun lambda invoke $output  --function-name $stackName --invocation-type RequestResponse && \
cat $output && \
grep 'Starting hander' $output && \
grep '"ZAP_TEST":"something"' $output > /dev/null && 
grep '"functionName":"zap-test"' $output > /dev/null && 
grep  '"requestId":"[a-z0-9-]{36}"' $output > /dev/null

let r=$?
echo "Exit with r"
rm -f $output
exit $r