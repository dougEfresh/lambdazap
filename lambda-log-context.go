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
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LambdaField type alias
type LambdaField int

// Enums for fields
const (
	AwsRequestID LambdaField = iota
	CognitoIdentityID
	CognitoIdentityPoolID
	InstallationID
	AppTitle
	AppVersionCode
	AppPackageName
	InvokeFuntionArn
	FunctionName
	FunctionVersion
	LogGroupName
	LogStreamName
	MemoryLimitInMB
	END
)

const staticStartIndex = 8

// DefaultNames of fields
var DefaultNames = []string{
	AwsRequestID:          "requestId",
	FunctionName:          "functionName",
	FunctionVersion:       "functionVersion",
	LogGroupName:          "logGroupName",
	LogStreamName:         "logStreamName",
	InvokeFuntionArn:      "arn",
	CognitoIdentityID:     "cognitoIdentityId",
	CognitoIdentityPoolID: "cognitoIdentityPoolId",
	InstallationID:        "installationId",
	AppTitle:              "appTitle",
	AppVersionCode:        "appVersionCode",
	AppPackageName:        "appPackageName",
	MemoryLimitInMB:       "memoryLimitInMB",
}

// An Option configures a Logger.
type Option interface {
	apply(*LambdaLogContext)
}

type optionFunc func(*LambdaLogContext)

func (f optionFunc) apply(lc *LambdaLogContext) {
	f(lc)
}

// CustomNames for fields
func CustomNames(n map[LambdaField]string) Option {
	return optionFunc(func(lc *LambdaLogContext) {
		lc.customNames = n
	})
}

// CustomValues override context values. if you return an error the default ContextValue will be used.
func CustomValues(c ContextValuer) Option {
	return optionFunc(func(lc *LambdaLogContext) {
		lc.customBuilder = c
	})
}

// ProcessNonContextFields when calling ContextValues include or not include
// Values that are not part of the lambda context
func ProcessNonContextFields(b bool) Option {
	return optionFunc(func(lc *LambdaLogContext) {
		lc.processNonContextValues = b
	})
}

// ContextValuer Control how you get the value from a field and context
type ContextValuer interface {
	ContextValue(ctx *lambdacontext.LambdaContext, f LambdaField) (string, error)
}

// LambdaLogContext structure
type LambdaLogContext struct {
	customBuilder           ContextValuer
	customNames             map[LambdaField]string
	ctxFields               map[int]int
	staticFields            []zapcore.Field
	fields                  []zapcore.Field
	processNonContextValues bool
}

var emptyvalues = make([]zapcore.Field, 0)

// New Create a new LambdaLogContext
func New(options ...Option) *LambdaLogContext {
	l := &LambdaLogContext{processNonContextValues: true}
	if len(options) > 0 {
		l.WithOptions(options...)
	}
	l.fields = make([]zapcore.Field, 0)
	l.staticFields = make([]zapcore.Field, 0)
	l.ctxFields = make(map[int]int)
	return l
}

// WithOptions add these options to the context.
func (lc *LambdaLogContext) WithOptions(opts ...Option) *LambdaLogContext {
	for _, v := range opts {
		v.apply(lc)
	}
	return lc
}

// WithBasic Add basic logging context
// See ...
func (lc *LambdaLogContext) WithBasic() *LambdaLogContext {
	return lc.With(AwsRequestID, FunctionName, FunctionVersion, InvokeFuntionArn, LogGroupName, LogStreamName)
}

// WithAll Add all fields to  logging context
// See ...
func (lc *LambdaLogContext) WithAll() *LambdaLogContext {
	return lc.WithBasic().With(CognitoIdentityID, CognitoIdentityPoolID, InstallationID, AppTitle, AppVersionCode, AppPackageName, MemoryLimitInMB)
}

func (lc *LambdaLogContext) getName(l LambdaField) string {
	n, ok := lc.customNames[l]
	if !ok {
		return DefaultNames[l]
	}
	return n
}

var dummyCtx = &lambdacontext.LambdaContext{}

// With Add these fields to context Add static fields if processNonContextValues is true
func (lc *LambdaLogContext) With(fields ...LambdaField) *LambdaLogContext {
	var ctxFieldIndex = len(lc.fields)

	for _, f := range fields {
		//		fmt.Fprintln(os.Stderr, "Adding field ", f, " index ", ctxFieldIndex)
		if int(f) >= staticStartIndex {
			field := zap.String(lc.getName(f), Extract(dummyCtx, f))
			if f == MemoryLimitInMB {
				memAsStr = fmt.Sprintf("%d", lambdacontext.MemoryLimitInMB)
				field = zap.Int(lc.getName(f), lambdacontext.MemoryLimitInMB)
			}
			lc.staticFields = append(lc.staticFields, field)
			if lc.processNonContextValues {
				lc.ctxFields[int(f)] = ctxFieldIndex
				lc.fields = append(lc.fields, field)
				ctxFieldIndex++
			}
		} else {
			lc.ctxFields[int(f)] = ctxFieldIndex
			lc.fields = append(lc.fields, zap.String(lc.getName(f), ""))
			ctxFieldIndex++
		}
	}
	return lc
}

// NonContextValues e.g. lambdacontext.FunctionName or os.Getenv
func (lc *LambdaLogContext) NonContextValues() []zapcore.Field {
	return lc.staticFields
}

// ContextValue get the context value for a field
func (lc *LambdaLogContext) ContextValue(ctx *lambdacontext.LambdaContext, f LambdaField) string {
	if lc.customBuilder != nil {
		v, err := lc.customBuilder.ContextValue(ctx, f)
		if err == nil {
			return v
		}
	}
	return Extract(ctx, f)
}

// ContextValues for the lambda context.
func (lc *LambdaLogContext) ContextValues(ctx context.Context) []zapcore.Field {
	lcv, ok := lambdacontext.FromContext(ctx)
	if len(lc.ctxFields) == 0 || !ok {
		return emptyvalues
	}
	for k, v := range lc.ctxFields {
		if k < int(END) {
			if LambdaField(k) < staticStartIndex {
				lc.fields[v].String = lc.ContextValue(lcv, LambdaField(k))
			}
		}
		if k >= 200 {
			//Custom fields start at 200
			lc.fields[v].String = lcv.ClientContext.Custom[lc.fields[v].Key]
		}
	}
	return lc.fields
}

// WithEnv Add Env from os.Getenv
func (lc *LambdaLogContext) WithEnv(names ...string) *LambdaLogContext {
	var start = len(lc.fields)
	for _, n := range names {
		lc.ctxFields[start+100] = start
		f := zap.String(n, os.Getenv(n))
		lc.staticFields = append(lc.staticFields, f)
		lc.fields = append(lc.fields, f)
		start++
	}
	return lc
}

// WithCustom Add names from lambdacontext.ClientContext.Custom
func (lc *LambdaLogContext) WithCustom(names ...string) *LambdaLogContext {
	var start = len(lc.fields)
	for _, n := range names {
		lc.ctxFields[start+200] = start
		lc.fields = append(lc.fields, zap.String(n, ""))
		start++
	}
	return lc
}

var memAsStr = ""

// Extract a field from lambda context
func Extract(ctx *lambdacontext.LambdaContext, field LambdaField) string {
	switch field {
	case AwsRequestID:
		return ctx.AwsRequestID
	case FunctionName:
		return lambdacontext.FunctionName
	case FunctionVersion:
		return lambdacontext.FunctionVersion
	case InvokeFuntionArn:
		return ctx.InvokedFunctionArn
	case LogGroupName:
		return lambdacontext.LogGroupName
	case LogStreamName:
		return lambdacontext.LogStreamName
	case CognitoIdentityID:
		return ctx.Identity.CognitoIdentityID
	case CognitoIdentityPoolID:
		return ctx.Identity.CognitoIdentityPoolID
	case InstallationID:
		return ctx.ClientContext.Client.InstallationID
	case AppTitle:
		return ctx.ClientContext.Client.AppTitle
	case AppVersionCode:
		return ctx.ClientContext.Client.AppVersionCode
	case AppPackageName:
		return ctx.ClientContext.Client.AppPackageName
	case MemoryLimitInMB:
		return memAsStr
	default:
		return ""
	}
}
