// Copyright © 2018 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lambdazap

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/stretchr/testify/assert"
)

type testWriter struct {
	value map[string]interface{}
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	json.Unmarshal(p, &tw.value)
	delete(tw.value, "ts")
	delete(tw.value, "level")
	return len(p), nil
}

func (tw *testWriter) Sync() error {
	return nil
}

func getLogger() (*zap.Logger, *testWriter) {
	cfg := zap.NewProductionEncoderConfig()
	en := zapcore.NewJSONEncoder(cfg)
	tw := &testWriter{}
	ws := zapcore.AddSync(tw)
	core := zapcore.NewCore(en, ws, zap.InfoLevel)
	return zap.New(core), tw
}

func TestNoContext(t *testing.T) {
	lf := New()
	logger, tw := getLogger()
	logger = logger.With(lf.NonContextValues()...)
	logger.Info("test", lf.ContextValues(context.TODO())...)
	logger.Sync()
	//contextValues := NewKitContext().With(AwsRequestID).ContextValues(context.TODO())
	if len(tw.value) != 1 {
		t.Fatal("Should only be 3 key in log output ", tw.value)
	}
	if tw.value["msg"] != "test" {
		t.Fatal("mismatch log output ", tw.value)
	}
}

func TestOneContext(t *testing.T) {
	defer func() {
		reset()
	}()
	setStatics()
	lc, cf := getContext()
	defer cf()
	lf := New().With(AwsRequestID)
	logger, tw := getLogger()
	//t.Log(contextValues)
	logger.Info("test", lf.ContextValues(lc)...)
	logger.Sync()
	//t.Log(tw.value)
	if len(tw.value) != 2 {
		t.Fatal("Should only be 2 keys in log output ", tw.value)
	}
	assert.Equal(t, "dummyid", tw.value["requestId"])
}

func TestBasicContext(t *testing.T) {
	defer func() {
		reset()
	}()
	setStatics()
	lc, cf := getContext()
	defer cf()
	lf := New().WithBasic()
	logger, tw := getLogger()
	contextValues := lf.ContextValues(lc)
	//t.Log(contextValues)
	logger = logger.With(lf.NonContextValues()...)
	logger.Info("test", contextValues...)
	logger.Sync()
	if len(tw.value) != 7 {
		t.Fatal("Should only 7 keys in output ", tw.value, len(tw.value))
	}

	assert.Equal(t, "dummyid", tw.value["requestId"])
	assert.Equal(t, "dummyfunction", tw.value["functionName"])
	assert.Equal(t, "dummyversion", tw.value["functionVersion"])
	assert.Equal(t, "dummylog", tw.value["logGroupName"])
	assert.Equal(t, "dummystream", tw.value["logStreamName"])
	assert.Equal(t, "dummyarn", tw.value["arn"])

}

func TestAllContext(t *testing.T) {
	defer func() {
		reset()
	}()
	setStatics()
	lc, cf := getContext()
	defer cf()
	lf := New().WithAll()
	logger, tw := getLogger()
	contextValues := lf.ContextValues(lc)
	//t.Log(contextValues)
	logger = logger.With(lf.NonContextValues()...)
	logger.Info("test", contextValues...)
	logger.Sync()
	//t.Log("value", tw.value, len(tw.value))
	if len(tw.value) != 14 {
		t.Fatal("Should only 14 keys in output ", tw.value, len(tw.value))
	}

	assert.Equal(t, "dummyid", tw.value["requestId"])
	assert.Equal(t, "dummyfunction", tw.value["functionName"])
	assert.Equal(t, "dummyversion", tw.value["functionVersion"])
	assert.Equal(t, "dummylog", tw.value["logGroupName"])
	assert.Equal(t, "dummystream", tw.value["logStreamName"])
	assert.Equal(t, "dummyarn", tw.value["arn"])
	assert.Equal(t, float64(128), tw.value["memoryLimitInMB"])
	assert.Equal(t, "dummyinstallid", tw.value["installationId"])
	assert.Equal(t, "dummytitle", tw.value["appTitle"])
	assert.Equal(t, "dummyname", tw.value["appPackageName"])
	assert.Equal(t, "dummycode", tw.value["appVersionCode"])
	assert.Equal(t, "dummypool", tw.value["cognitoIdentityPoolId"])
	assert.Equal(t, "dummyident", tw.value["cognitoIdentityId"])
}

func TestEnvContext(t *testing.T) {
	defer func() {
		reset()
	}()
	setStatics()
	lc, cf := getContext()
	defer cf()
	lf := New().With(AwsRequestID).With(FunctionName).With(CognitoIdentityID).WithEnv("SHELL").WithCustom("custom1")
	logger, tw := getLogger()
	contextValues := lf.ContextValues(lc)
	//t.Log(contextValues)
	logger = logger.With(lf.NonContextValues()...)
	logger.Info("test", contextValues...)
	logger.Sync()
	if len(tw.value) != 6 {
		t.Fatal("Should only be 6 keys in log output ", tw.value)
	}
	assert.Equal(t, "dummyid", tw.value["requestId"])
	assert.Contains(t, tw.value["SHELL"], "/bin")
	assert.Equal(t, "dummycustom1", tw.value["custom1"])
	assert.Equal(t, "dummyfunction", tw.value["functionName"])
	assert.Equal(t, "dummyident", tw.value["cognitoIdentityId"])
}

var cNames = map[LambdaField]string{
	AwsRequestID: "rId",
}

type custValuer struct{}

func (c *custValuer) ContextValue(ctx *lambdacontext.LambdaContext, f LambdaField) (string, error) {
	if f == AwsRequestID {
		return "customId", nil
	}

	return "", fmt.Errorf("use default")
}

func TestCustomContext(t *testing.T) {
	defer func() {
		reset()
	}()
	setStatics()
	lc, cf := getContext()
	defer cf()
	lf := New(CustomValues(&custValuer{}), ProcessNonContextFields(true)).With(AwsRequestID).With(FunctionName).With(CognitoIdentityID).WithEnv("SHELL").WithCustom("custom1")
	logger, tw := getLogger()
	contextValues := lf.ContextValues(lc)
	contextValues = append(contextValues, lf.NonContextValues()...)
	//t.Log(contextValues)
	//logger = logger.With(lf.NonContextValues()...).With(contextValues...)
	logger.Info("test", contextValues...)
	logger.Sync()
	if len(tw.value) != 6 {
		t.Fatal("Should only be 6 keys in log output ", tw.value)
	}
	assert.Equal(t, "customId", tw.value["requestId"])
	assert.Contains(t, tw.value["SHELL"], "/bin")
}

func TestCustomName(t *testing.T) {
	defer func() {
		reset()
	}()
	setStatics()
	lc, cf := getContext()
	defer cf()
	lf := New(CustomValues(&custValuer{}), CustomNames(cNames)).With(AwsRequestID).With(FunctionName).With(CognitoIdentityID).WithEnv("SHELL").WithCustom("custom1")
	logger, tw := getLogger()
	contextValues := lf.ContextValues(lc)
	//t.Log(contextValues)
	logger = logger.With(lf.NonContextValues()...)
	logger.Info("test", contextValues...)
	logger.Sync()
	if len(tw.value) != 6 {
		t.Fatal("Should only be 6 keys in log output ", tw.value)
	}
	assert.Equal(t, "customId", tw.value["rId"])
	assert.Contains(t, tw.value["SHELL"], "/bin")
}

func setStatics() {
	lambdacontext.FunctionName = "dummyfunction"
	lambdacontext.FunctionVersion = "dummyversion"
	lambdacontext.LogGroupName = "dummylog"
	lambdacontext.LogStreamName = "dummystream"
	lambdacontext.MemoryLimitInMB = 128
}

func getContext() (context.Context, context.CancelFunc) {
	deadline := time.Unix(req.Deadline.Seconds, req.Deadline.Nanos).UTC()
	invokeContext, cancel := context.WithDeadline(context.Background(), deadline)
	invokeContext = lambdacontext.NewContext(invokeContext, lc)
	if len(req.ClientContext) > 0 {
		if err := json.Unmarshal(req.ClientContext, &lc.ClientContext); err != nil {
			panic(fmt.Sprintf("error parsing %s", err))
		}
	}
	return lambdacontext.NewContext(invokeContext, lc), cancel
}

var req = &messages.InvokeRequest{
	CognitoIdentityId:     "dummyident",
	CognitoIdentityPoolId: "dummypool",
	ClientContext: []byte(`{
		"Client": {
			"app_title": "dummytitle",
			"installation_id": "dummyinstallid",
			"app_version_code": "dummycode",
			"app_package_name": "dummyname"
		},
		"env": {
			"env1": "1",
			"env2": "2"
		},
		"custom": {
			"custom1": "dummycustom1",
			"custom2": "dummycustom2"
		}
	}`),
	RequestId:          "dummyid",
	InvokedFunctionArn: "dummyarn",
}
var lc = &lambdacontext.LambdaContext{
	AwsRequestID:       req.RequestId,
	InvokedFunctionArn: req.InvokedFunctionArn,
	Identity: lambdacontext.CognitoIdentity{
		CognitoIdentityID:     req.CognitoIdentityId,
		CognitoIdentityPoolID: req.CognitoIdentityPoolId,
	},
}

func reset() {
	lambdacontext.FunctionName = ""
	lambdacontext.MemoryLimitInMB = 0
	lambdacontext.FunctionVersion = ""
	lambdacontext.LogGroupName = ""
	lambdacontext.LogStreamName = ""
}
