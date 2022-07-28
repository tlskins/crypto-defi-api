package api

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/davecgh/go-spew/spew"
)

type LambdaResp events.APIGatewayProxyResponse

type LambdaReq events.APIGatewayProxyRequest

type LambdaHandlerFunc func(context.Context, *LambdaReq) (resp LambdaResp, err error)

func BuildLambdaResponse(data interface{}, allowedOrigin string) (resp LambdaResp, err error) {
	var buf bytes.Buffer
	var body []byte
	if body, err = json.Marshal(data); err != nil {
		return
	}
	json.HTMLEscape(&buf, body)

	resp = LambdaResp{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      allowedOrigin,
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "OPTIONS,POST,GET",
			"Access-Control-Allow-Headers":     "Content-Type",
		},
	}
	return
}

func BuildLambdaCSVResponse(csvData [][]string, allowedOrigin string) (resp LambdaResp) {
	buf := &bytes.Buffer{} // creates IO Writer
	writer := csv.NewWriter(buf)
	writer.WriteAll(csvData)
	writer.Flush()

	now := time.Now()
	resp = LambdaResp{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":                     "text/csv",
			"Access-Control-Allow-Origin":      allowedOrigin,
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "OPTIONS,POST,GET",
			"Access-Control-Allow-Headers":     "Content-Type",
			"Content-Disposition":              fmt.Sprintf("attachment;filename=hungry_shipments_%s.csv", now.Format("01-02-2006-15-14")),
		},
	}
	return
}

func Parse(req *LambdaReq, out interface{}) {
	var jsonBytes []byte
	var err error
	if req.Body == "" {
		if jsonBytes, err = json.Marshal(req.QueryStringParameters); err != nil {
			Abort(http.StatusUnprocessableEntity, err)
		}
	} else {
		jsonBytes = []byte(req.Body)
	}
	if err = json.Unmarshal(jsonBytes, out); err != nil {
		Abort(http.StatusUnprocessableEntity, err)
	}
}

type Validator interface {
	Validate() error
}

func ParseAndValidate(req *LambdaReq, out Validator) {
	Parse(req, out)
	if err := out.Validate(); err != nil {
		fmt.Printf("err caught in ParseAndValidate %v\n", err)
		Abort(http.StatusUnprocessableEntity, err)
	}
}

type Responder struct {
	OriginStr string
}

func (r Responder) corsHeaders(headers map[string]string) map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":      r.Origin(headers),
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "OPTIONS,POST,GET",
		"Access-Control-Allow-Headers":     "Content-Type",
	}
}

func (r Responder) Origin(headers map[string]string) (origin string) {
	origins := strings.Split(r.OriginStr, ",")
	origin = origins[0]
	for _, str := range origins {
		if str == headers["origin"] {
			origin = str
		}
	}
	return
}

// Fail returns an internal server error with the error message
func (r Responder) Fail(headers map[string]string, msg string, status int) (LambdaResp, error) {
	e := make(map[string]string, 0)
	e["message"] = msg

	// We don't need to worry about this error,
	// as we're controlling the input.
	body, _ := json.Marshal(e)

	return LambdaResp{
		Body:       string(body),
		Headers:    r.corsHeaders(headers),
		StatusCode: status,
	}, nil
}

// Success returns a valid response
func (r Responder) Success(headers map[string]string, data interface{}, status int) (LambdaResp, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return r.Fail(headers, err.Error(), http.StatusInternalServerError)
	}

	return LambdaResp{
		Body:       string(body),
		Headers:    r.corsHeaders(headers),
		StatusCode: status,
	}, nil
}

func (responder Responder) HandleRequest(handle func(context.Context, *LambdaReq) (LambdaResp, error)) func(context.Context, *LambdaReq) (LambdaResp, error) {
	return func(ctx context.Context, req *LambdaReq) (resp LambdaResp, err error) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(Error); ok {
					resp, err = responder.Fail(req.Headers, e.String(), e.Code)
				} else if e, ok := r.(error); ok {
					resp, err = responder.Fail(req.Headers, e.Error(), http.StatusInternalServerError)
				} else {
					resp, err = responder.Fail(req.Headers, "unknown error", http.StatusInternalServerError)
				}
				spew.Dump(err)
				debug.PrintStack()
			}
		}()
		resp, err = handle(ctx, req)
		return
	}
}
