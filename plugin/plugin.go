// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Args struct {
	Pipeline
	PluginInputParams
	PluginConfigParams
}

type PluginConfigParams struct {
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`
}

type PluginInputParams struct {
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
}

type PluginProcessingInfo struct {
	AuthUser string
	AuthPass string
	//BodyBytes       *bytes.Buffer
	BodyIoReader     io.Reader
	HttpReq          *http.Request
	TimeOutDuration  time.Duration
	httpClient       *http.Client
	httpResponse     *http.Response
	isConnectionOpen bool
}

type PluginExecResultsCard struct {
	ResponseStatus  int    `json:"RESPONSE_STATUS"`
	ResponseContent string `json:"RESPONSE_CONTENT"`
	ResponseHeaders string `json:"RESPONSE_HEADERS"`
	ResponseFile    string `json:"RESPONSE_FILE"`
}

type Plugin struct {
	Args
	PluginProcessingInfo
	PluginExecResultsCard
}

//type Plugin = Plugin

func Exec(ctx context.Context, args Args) error {

	Plugin := GetNewPlugin(args)
	err := Plugin.Run()
	if err != nil {
		return err
	}
	return nil
}

func GetNewPlugin(args Args) *Plugin {
	return &Plugin{
		Args: args,
	}
}

func (p *Plugin) Init() error {
	return nil
}

func (p *Plugin) DeInit() error {
	if p.isConnectionOpen {
		p.httpResponse.Body.Close()
	}

	fmt.Println("DeInit() called")
	return nil
}

func (p *Plugin) Run() error {

	err := p.ValidateArgs()
	if err != nil {
		log.Println("ValidateArgs failed err == ", err.Error())
	}

	err = p.DoRequest()
	if err != nil {
		log.Println("DoRequest failed err == ", err.Error())
	}

	err = p.StoreHttpResponseResults()
	if err != nil {
		return errors.New("error storing response results")
	}
	return nil
}

func (p *Plugin) CreateNewHttpRequest() error {

	p.SetTimeout()

	ctx, cancel := context.WithTimeout(context.Background(), p.TimeOutDuration)
	defer cancel()

	var err error

	p.HttpReq, err = http.NewRequestWithContext(ctx, p.HttpMethod, p.Url, p.BodyIoReader)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	p.HttpReq, err = http.NewRequest(p.HttpMethod, p.Url, p.BodyIoReader)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err.Error())
	}
	return nil
}

func (p *Plugin) DoRequest() error {

	err := p.CreateNewHttpRequest()
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	p.SetHeaders()
	p.SetAuthBasic()
	p.CreateHttpClient()
	p.SetIsHonorSsl()

	p.httpResponse, err = p.httpClient.Do(p.HttpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("request timed out after %v", p.TimeOutDuration)
		}
		return fmt.Errorf("error sending request: %v", err)
	}
	p.isConnectionOpen = true

	if p.ResponseStatus < 200 && p.ResponseStatus > 299 {
		return errors.New("bad response status")
	}

	return nil
}

func (p *Plugin) StoreHttpResponseResults() error {

	headers := make([]string, 0, len(p.httpResponse.Header))

	p.ResponseStatus = p.httpResponse.StatusCode
	for key, values := range p.httpResponse.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", key, strings.Join(values, ",")))
	}
	p.ResponseHeaders = strings.Join(headers, "\n")

	switch {
	case len(p.OutputFile) > 0:
		err := p.WriteResponseToFile()
		if err != nil {
			return err
		}
	default:
		err := p.SetResponseBody()
		if err != nil {
			return err
		}
	}
	card := PluginExecResultsCard{
		ResponseStatus:  p.ResponseStatus,
		ResponseContent: p.ResponseContent,
		ResponseHeaders: p.ResponseHeaders,
		ResponseFile:    p.OutputFile,
	}

	writeCard(StdOut, Schema, card)
	return nil
}

func (p *Plugin) WriteResponseToFile() error {

	outFile, err := os.Create(p.OutputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, p.httpResponse.Body)
	if err != nil {
		return fmt.Errorf("error saving response to file: %v", err)
	}

	return nil
}

func (p *Plugin) SetResponseBody() error {
	bodyBytes, err := ioutil.ReadAll(p.httpResponse.Body)
	if err != nil {
		fmt.Println("error reading response body ", err.Error())
		return err
	}
	p.ResponseContent = string(bodyBytes)
	return nil
}

func (p *Plugin) CreateHttpClient() {
	p.httpClient = &http.Client{
		Timeout: p.TimeOutDuration,
	}
}

func (p *Plugin) SetHttpResponseEnvVars() {
	fmt.Printf("Response status: %s\n", p.httpResponse.Status)
	fmt.Println("Response headers:")
	for key, values := range p.httpResponse.Header {
		fmt.Printf("%s: %s\n", key, strings.Join(values, ","))
	}
}

func (p *Plugin) SetIsHonorSsl() {
	if p.IgnoreSsl {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		p.httpClient.Transport = transport
	}
}

func (p *Plugin) SetAuthBasic() {

	if p.AuthBasic != "" {
		return
	}

	if p.AuthUser != "" && p.AuthPass != "" {
		p.HttpReq.SetBasicAuth(p.AuthUser, p.AuthPass)
	}
}

func (p *Plugin) SetTimeout() {
	if p.Timeout == 0 {
		p.TimeOutDuration = 60 * time.Second
		return
	}
	p.TimeOutDuration = time.Duration(p.Timeout) * time.Second
}

func (p *Plugin) SetHeaders() {
	headersStr := p.Headers
	if headersStr != "" {
		headers := strings.Split(headersStr, ",")
		for _, header := range headers {
			kvPair := strings.SplitN(header, ":", 2)
			p.HttpReq.Header.Set(strings.TrimSpace(kvPair[0]), strings.TrimSpace(kvPair[1]))
		}
	}
}

func (p *Plugin) ValidateArgs() error {

	if p.ValidateUrl() != nil {
		fmt.Println("bad url")
		return errors.New("url is required")
	}

	if p.ValidateHttpMethod(p.HttpMethod) != nil {
		fmt.Println("invalid http_method")
		return errors.New("invalid http_method")
	}

	if p.ValidateHeader(p.Headers) != nil {
		fmt.Println("malformed headers")
		return errors.New("malformed headers")
	}

	if p.ValidateRequestBody() != nil {
		fmt.Println("request_body is required")
		return errors.New("request_body is required")
	}

	if p.ValidateAuthBasic() != nil {
		fmt.Println("auth_basic info not good")
		return errors.New("auth_basic info not good")
	}

	return nil
}

func (p *Plugin) ValidateAuthBasic() error {

	if p.AuthBasic == "" {
		return nil
	}

	authBasic := p.AuthBasic
	userPassInfo := strings.Split(authBasic, ":")

	if len(userPassInfo) != 2 {
		return errors.New("invalid auth_basic format")
	}

	p.AuthUser = userPassInfo[0]
	p.AuthPass = userPassInfo[1]

	return nil
}

func (p *Plugin) ValidateRequestBody() error {

	bodyStr := p.RequestBody
	method := p.HttpMethod

	if bodyStr != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
		p.BodyIoReader = strings.NewReader(p.RequestBody)
	} else {
		p.BodyIoReader = nil
	}

	return nil
}

func (p *Plugin) ValidateUrl() error {
	if p.Url == "" {
		return errors.New("url is required")
	}
	return nil
}

func (p *Plugin) ValidateHttpMethod(httpMethod string) error {

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

func (p *Plugin) ValidateHeader(headerStr string) error {

	headersList := strings.Split(headerStr, ",")

	for i, headerItem := range headersList {

		headerItem = strings.TrimSpace(headerItem)
		if i == 0 && headerItem == "" {
			fmt.Println(`if i == 0 && headerItem == ""`)
			return errors.New("malformed header: empty header")
		}

		kvPair := strings.SplitN(headerItem, ":", 2)
		if len(kvPair) != 2 {
			fmt.Println(`if len(kvPair) != 2`)
			return fmt.Errorf("malformed header: '%s' (missing colon)", headerItem)
		}

		key := strings.TrimSpace(kvPair[0])
		if key == "" {
			fmt.Println(`if key == ""`)
			return fmt.Errorf("malformed header: '%s' (empty header name)", headerItem)
		}

		value := strings.TrimSpace(kvPair[1])
		if value == "" {
			fmt.Println(`if value == ""`)
			return fmt.Errorf("malformed header: '%s' (empty header value)", headerItem)
		}
	}

	return nil
}

const (
	Schema = "https://drone.github.io/drone-jira/card.json"
	StdOut = "/dev/stdout"
)

//
//
