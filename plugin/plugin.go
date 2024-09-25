// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Plugin Params

	Url                string `envconfig:"PLUGIN_URL"`
	HttpMethod         string `envconfig:"PLUGIN_HTTP_METHOD"`
	Headers            string `envconfig:"PLUGIN_HEADERS"`
	RequestBody        string `envconfig:"PLUGIN_REQUEST_BODY"`
	AuthBasic          string `envconfig:"PLUGIN_AUTH_BASIC"`
	AuthCert           string `envconfig:"PLUGIN_AUTH_CERT"`
	ValidResponseCodes string `envconfig:"PLUGIN_VALID_RESPONSE_CODES"`
	ValidResponseBody  string `envconfig:"PLUGIN_VALID_RESPONSE_BODY"`
	Timeout            int    `envconfig:"PLUGIN_TIMEOUT"`
	IgnoreSsl          bool   `envconfig:"PLUGIN_IGNORE_SSL"`
	Proxy              string `envconfig:"PLUGIN_PROXY"`
	OutputFile         string `envconfig:"PLUGIN_OUTPUT_FILE"`
	LogResponse        bool   `envconfig:"PLUGIN_LOG_RESPONSE"`
	Quiet              bool   `envconfig:"PLUGIN_QUIET"`
	WrapAsMultipart    bool   `envconfig:"PLUGIN_WRAP_AS_MULTIPART"`
	SslCertPath        string `envconfig:"PLUGIN_SSL_CERT_PATH"`

	//
	AuthUser        string
	AuthPass        string
	BodyBytes       *bytes.Buffer
	HttpReq         *http.Request
	TimeOutDuration time.Duration
	HttpClient      *http.Client
	HttpResponse    *http.Response
}

/*
	RESPONSE_STATUS -	The HTTP status code returned by the server.
	RESPONSE_CONTENT - The content of the response body (if not saved to a file i.e. when PLUGIN_OUTPUT_FILE is set).
	RESPONSE_HEADERS - The headers returned in the HTTP response.
	RESPONSE_FILE -	If output_file (PLUGIN_OUTPUT_FILE) is set, the file path where the response is saved.
*/

type OutputEnvVars struct {
	ResponseStatus  int    `json:"RESPONSE_STATUS"`
	ResponseContent string `json:"RESPONSE_CONTENT"`
	ResponseHeaders string `json:"RESPONSE_HEADERS"`
	ResponseFile    string `json:"RESPONSE_FILE"`
}

func Exec(ctx context.Context, args Args) error {
	httpRequestPlugin := GetNewHttpRequestPlugin(args)
	err := httpRequestPlugin.Run()
	if err != nil {
		return err
	}
	return nil
}

type HttpRequestPlugin struct {
	Args
}

func GetNewHttpRequestPlugin(args Args) *HttpRequestPlugin {
	return &HttpRequestPlugin{
		Args: args,
	}
}

func (p *HttpRequestPlugin) Run() error {
	err := p.ValidateArgs()
	if err != nil {
		log.Println("ValidateArgs failed err == ", err.Error())
	}

	err = p.DoRequest()
	if err != nil {
		log.Println("DoRequest failed err == ", err.Error())
	}

	p.SetOutputEnvVars()
	return nil
}

func (p *HttpRequestPlugin) DoRequest() error {

	err := p.CreateNewHttpRequest()
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	p.SetHeaders()
	p.SetAuthBasic()
	p.SetTimeout()
	p.CreateHttpClient()
	p.SetIsHonorSsl()

	p.HttpResponse, err = p.HttpClient.Do(p.HttpReq)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer p.HttpResponse.Body.Close()

	p.SetHttpResponseEnvVars()
	return nil

}

func (p *HttpRequestPlugin) CreateNewHttpRequest() error {
	var err error
	p.HttpReq, err = http.NewRequest(p.HttpMethod, p.Url, p.BodyBytes)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	return nil
}

func (p *HttpRequestPlugin) CreateHttpClient() {
	p.HttpClient = &http.Client{
		Timeout: p.TimeOutDuration,
	}
}

func (p *HttpRequestPlugin) SetHttpResponseEnvVars() {
	fmt.Printf("Response status: %s\n", p.HttpResponse.Status)
	fmt.Println("Response headers:")
	for key, values := range p.HttpResponse.Header {
		fmt.Printf("%s: %s\n", key, strings.Join(values, ","))
	}
}

func (p *HttpRequestPlugin) SetIsHonorSsl() {
	if p.IgnoreSsl {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		p.HttpClient.Transport = transport
	}
}

func (p *HttpRequestPlugin) SetAuthBasic() {
	if p.AuthUser != "" && p.AuthPass != "" {
		p.HttpReq.SetBasicAuth(p.AuthUser, p.AuthPass)
	}
}

func (p *HttpRequestPlugin) SetTimeout() {
	timeout := time.Duration(60) * time.Second
	if p.Timeout != 0 {
		timeout = time.Duration(p.Timeout) * time.Second
	}
	p.TimeOutDuration = timeout
}

func (p *HttpRequestPlugin) SetHeaders() {
	headersStr := p.Headers
	if headersStr != "" {
		headers := strings.Split(headersStr, ",")
		for _, header := range headers {
			kvPair := strings.SplitN(header, ":", 2)
			p.HttpReq.Header.Set(strings.TrimSpace(kvPair[0]), strings.TrimSpace(kvPair[1]))
		}
	}
}

func (p *HttpRequestPlugin) SetOutputEnvVars() {

	outputEnvVars := OutputEnvVars{
		ResponseStatus:  200,
		ResponseContent: "",
		ResponseHeaders: "",
		ResponseFile:    "",
	}

	// Write the output environment variables to the standard output
	writeCard("/dev/stdout", "drone", outputEnvVars)
}

func (p *HttpRequestPlugin) ValidateArgs() error {

	if p.ValidateUrl() != nil {
		return errors.New("url is required")
	}

	if p.ValidateHttpMethod(p.HttpMethod) != nil {
		return errors.New("invalid http_method")
	}

	if p.ValidateHeader(p.Headers) != nil {
		return errors.New("malformed headers")
	}

	if p.ValidateRequestBody() != nil {
		return errors.New("request_body is required")
	}

	return nil
}

func (p *HttpRequestPlugin) ValidateAuthBasic() error {

	authBasic := p.AuthBasic

	if authBasic == "" {
		return nil
	}

	const prefix = "Authorization: Basic "
	if !strings.HasPrefix(authBasic, prefix) {
		return errors.New("invalid authorization header format")
	}

	b64Creds := strings.TrimPrefix(authBasic, prefix)

	decodedBytes, err := base64.StdEncoding.DecodeString(b64Creds)
	if err != nil {
		return errors.New("invalid base64 encoding")
	}
	decodedCredentials := string(decodedBytes)

	credentials := strings.SplitN(decodedCredentials, ":", 2)
	if len(credentials) != 2 {
		return errors.New("invalid credentials format, expected 'username:password'")
	}

	if len(credentials[0]) != 0 && len(credentials[1]) == 0 {
		return errors.New("username and password cannot be empty")
	} else if len(credentials[0]) == 0 && len(credentials[1]) != 0 {
		return errors.New("username and password cannot be empty")
	}

	p.AuthUser = credentials[0]
	p.AuthPass = credentials[1]

	return nil
}

func (p *HttpRequestPlugin) ValidateRequestBody() error {

	bodyStr := p.RequestBody
	method := p.HttpMethod

	if bodyStr != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
		p.BodyBytes = bytes.NewBuffer([]byte(bodyStr))
	} else {
		p.BodyBytes = nil
	}

	return nil // body can be empty in most cases
}

func (p *HttpRequestPlugin) ValidateUrl() error {
	if p.Url == "" {
		return errors.New("url is required")
	}
	return nil
}

func (p *HttpRequestPlugin) ValidateHttpMethod(httpMethod string) error {

	if httpMethod == "" {
		p.HttpMethod = "GET"
		return nil
	}

	httpMethod = strings.ToUpper(httpMethod)
	switch httpMethod {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "MKCOL":
		return nil
	default:
		return errors.New("invalid http_method")
	}
}

func (p *HttpRequestPlugin) ValidateHeader(headerStr string) error {

	headersList := strings.Split(headerStr, ",")

	for i, headerItem := range headersList {

		headerItem = strings.TrimSpace(headerItem)
		if i == 0 && headerItem == "" {
			return errors.New("malformed header: empty header")
		}

		kvPair := strings.SplitN(headerItem, ":", 2)
		if len(kvPair) != 2 {
			return fmt.Errorf("malformed header: '%s' (missing colon)", headerItem)
		}

		key := strings.TrimSpace(kvPair[0])
		if key == "" {
			return fmt.Errorf("malformed header: '%s' (empty header name)", headerItem)
		}

		value := strings.TrimSpace(kvPair[1])
		if value == "" {
			return fmt.Errorf("malformed header: '%s' (empty header value)", headerItem)
		}
	}

	return nil
}

//
//
