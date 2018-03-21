// Copyright © 2018 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dougEfresh/lambdazap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Create a new lambda log context and use RequestID... and a variable from environment
var lambdazapper = lambdazap.New().
	With(lambdazap.AwsRequestID, lambdazap.FunctionName, lambdazap.InvokeFunctionArn).
	WithEnv("ZAP_TEST")
var logger *zap.Logger

func init() {
	// Init the logger outside of the handler
	logger = getLogger()
}

type lambdaWrtier struct {
	value map[string]interface{}
}

var writer = &lambdaWrtier{}

func (w *lambdaWrtier) Write(p []byte) (n int, err error) {
	json.Unmarshal(p, &w.value)
	fmt.Fprintf(os.Stdout, "%s", string(p))
	return len(p), nil
}

func getLogger() *zap.Logger {
	en := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(en, zapcore.AddSync(writer), zap.InfoLevel)
	return zap.New(core)
}

// Handler for lambda
func Handler(ctx context.Context) (map[string]interface{}, error) {
	// defer logger.Sync()
	logger.Info("Starting hander with context values ", lambdazapper.ContextValues(ctx)...)
	logger.Sync() // Make sure to flush any buffers
	return writer.value, nil
}

func main() {
	lambda.Start(Handler)
}
