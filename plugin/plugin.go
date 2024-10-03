// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	ContentType        string `envconfig:"PLUGIN_CONTENT_TYPE"`
	RequestBody        string `envconfig:"PLUGIN_REQUEST_BODY"`
	AuthBasic          string `envconfig:"PLUGIN_AUTH_BASIC"`
	AuthCert           string `envconfig:"PLUGIN_AUTH_CERT"`
	ValidResponseCodes string `envconfig:"PLUGIN_VALID_RESPONSE_CODES"`
	ValidResponseBody  string `envconfig:"PLUGIN_VALID_RESPONSE_BODY"`
	Timeout            int    `envconfig:"PLUGIN_TIMEOUT"`
	IgnoreSsl          bool   `envconfig:"PLUGIN_IGNORE_SSL"`
	Proxy              string `envconfig:"PLUGIN_PROXY"`
	OutputFile         string `envconfig:"PLUGIN_OUTPUT_FILE"`
	AcceptType         string `envconfig:"PLUGIN_ACCEPT_TYPE"`
	LogResponse        bool   `envconfig:"PLUGIN_LOG_RESPONSE"`
	Quiet              bool   `envconfig:"PLUGIN_QUIET"`
	UploadFile         string `envconfig:"PLUGIN_UPLOAD_FILE"`
	MultiPartName      string `envconfig:"PLUGIN_MULTIPART_NAME"`
	WrapAsMultipart    bool   `envconfig:"PLUGIN_WRAP_AS_MULTIPART"`
	SslCertPath        string `envconfig:"PLUGIN_SSL_CERT_PATH"`
}

type PluginProcessingInfo struct {
	AuthUser                 string
	AuthPass                 string
	HttpRequestCancelContext context.CancelFunc
	BodyIoReader             io.Reader
	HttpReq                  *http.Request
	TimeOutDuration          time.Duration
	httpClient               *http.Client
	httpResponse             *http.Response
	httpResponseBodyBytes    []byte
	isConnectionOpen         bool
	proxyUrl                 *url.URL
	uploadFileAbsolutePath   string
	IsSuppressLogs           bool
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

func Exec(ctx context.Context, args Args) error {

	plugin := GetNewPlugin(args)

	_ = plugin.Init()

	err := plugin.Run()
	if err != nil {
		return err
	}

	err = plugin.DeInit()
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

	var err error

	if p.isConnectionOpen {
		p.isConnectionOpen = false
		err = p.httpResponse.Body.Close()
	}

	if p.HttpRequestCancelContext != nil {
		p.HttpRequestCancelContext()
		p.HttpRequestCancelContext = nil
	}

	if p.BodyIoReader != nil {
		if closer, ok := p.BodyIoReader.(io.Closer); ok {
			err = closer.Close()
		}
	}

	LogPrintln(p, "DeInit() called")
	return err
}

func (p *Plugin) Run() error {

	err := p.ValidateArgs()
	if err != nil {
		log.Println("ValidateArgs failed err == ", err.Error())
		return err
	}

	err = p.DoRequest()
	if err != nil {
		log.Println("DoRequest failed err == ", err.Error())
		return err
	}

	if p.httpResponse.StatusCode == http.StatusUnsupportedMediaType {
		return errors.New(fmt.Sprintf("http.StatusUnsupportedMediaType == %d", http.StatusUnsupportedMediaType))
	}

	err = p.StoreHttpResponseResults()
	if err != nil {
		return err
	}
	return nil
}

func (p *Plugin) CreateNewHttpRequest() error {

	p.SetTimeout()
	var ctx context.Context

	ctx, p.HttpRequestCancelContext = context.WithTimeout(context.Background(), p.TimeOutDuration)

	var err error

	p.HttpReq, err = http.NewRequestWithContext(ctx, p.HttpMethod, p.Url, p.BodyIoReader)
	if err != nil {
		return err
	}

	p.HttpReq.Header.Set(ContentType, ApplicationJson)

	if p.WrapAsMultipart || p.ContentType != "" {
		p.HttpReq.Header.Set(ContentType, p.ContentType)
	}

	if p.AcceptType != "" {
		p.HttpReq.Header.Set("Accept", p.AcceptType)
	}
	return nil
}

func (p *Plugin) DoRequest() error {

	err := p.CreateNewHttpRequest()
	if err != nil {
		return err
	}

	err = p.SetHeaders()
	if err != nil {
		return err
	}

	err = p.SetAuthBasic()
	if err != nil {
		return err
	}

	p.GetNewHttpClient()

	err = p.SetHttpConnectionParameters()
	if err != nil {
		return err
	}

	p.httpResponse, err = p.httpClient.Do(p.HttpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			LogPrintln(p, "request timed out")
		}
		return err
	}
	p.isConnectionOpen = true

	err = p.IsResponseStatusOk()
	if err != nil {
		return err
	}

	err = p.StoreHttpResponse()
	if err != nil {
		return err
	}

	err = p.CheckForValidResponseBody()
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) CheckForValidResponseBody() error {

	if len(p.ValidResponseBody) < 1 {
		return nil
	}

	if strings.Contains(p.ResponseContent, p.ValidResponseBody) {
		return nil
	}

	return errors.New("response body does not contain the expected string")
}

func (p *Plugin) IsResponseStatusOk() error {
	if p.ResponseStatus < 200 && p.ResponseStatus > 299 {
		return errors.New("bad response status")
	}
	return nil
}

func (p *Plugin) StoreHttpResponseResults() error {

	if p.LogResponse {
		LogPrintln(p, p.ResponseContent)
	}

	headers := make([]string, 0, len(p.httpResponse.Header))

	p.ResponseStatus = p.httpResponse.StatusCode
	for key, values := range p.httpResponse.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", key, strings.Join(values, ",")))
	}
	p.ResponseHeaders = strings.Join(headers, "\n")

	if len(p.OutputFile) > 0 {
		err := p.WriteResponseToFile()
		if err != nil {
			return err
		}
	}

	type EnvKvPair struct {
		Key   string
		Value interface{}
	}

	var kvPairs = []EnvKvPair{
		{"RESPONSE_STATUS", p.ResponseStatus},
		{"RESPONSE_CONTENT", p.ResponseContent},
		{"RESPONSE_HEADERS", p.ResponseHeaders},
		{"RESPONSE_FILE", p.OutputFile},
	}

	for _, kvPair := range kvPairs {
		err := WriteEnvToFile(kvPair.Key, kvPair.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) StoreHttpResponse() error {

	var err error

	p.httpResponseBodyBytes, err = io.ReadAll(p.httpResponse.Body)
	if err != nil {
		return err
	}

	p.ResponseContent = string(p.httpResponseBodyBytes)

	return nil
}

func (p *Plugin) WriteResponseToFile() error {

	outFile, err := os.Create(p.OutputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, strings.NewReader(p.ResponseContent))
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) GetNewHttpClient() {
	p.httpClient = &http.Client{
		Timeout: p.TimeOutDuration,
	}
}

func (p *Plugin) SetSslCert() {
	if p.AuthCert == "" || p.IgnoreSsl {
		return
	}
}

func (p *Plugin) SetAuthBasic() error {

	if p.AuthBasic == "" {
		return nil
	}

	if p.HttpReq == nil {
		return errors.New("SetAuthBasic http request is nil")
	}

	if p.AuthUser != "" && p.AuthPass != "" {
		p.HttpReq.SetBasicAuth(p.AuthUser, p.AuthPass)
	}
	return nil
}

func (p *Plugin) SetTimeout() {
	if p.Timeout == 0 {
		p.TimeOutDuration = 60 * time.Second
		return
	}
	p.TimeOutDuration = time.Duration(p.Timeout) * time.Second
}

func (p *Plugin) SetHeaders() error {

	headersStr := p.Headers

	if headersStr == "" {
		return nil
	}

	if p.HttpReq == nil {
		return errors.New("http request is nil")
	}

	if headersStr != "" {
		headers := strings.Split(headersStr, ",")
		for _, header := range headers {
			kvPair := strings.SplitN(header, ":", 2)
			p.HttpReq.Header.Set(strings.TrimSpace(kvPair[0]), strings.TrimSpace(kvPair[1]))
		}
	}

	return nil
}

func (p *Plugin) ValidateArgs() error {

	if p.ValidateUrl() != nil {
		LogPrintln(p, "bad url")
		return errors.New("url is required")
	}

	if p.ValidateHttpMethod(p.HttpMethod) != nil {
		LogPrintln(p, "invalid http_method")
		return errors.New("invalid http_method")
	}

	if p.ValidateHeader(p.Headers) != nil {
		LogPrintln(p, "malformed headers")
		return errors.New("malformed headers")
	}

	if p.ValidateRequestBody() != nil {
		LogPrintln(p, "request_body is required")
		return errors.New("request_body is required")
	}

	if p.ValidateAuthBasic() != nil {
		LogPrintln(p, "auth_basic info not good")
		return errors.New("auth_basic info not good")
	}

	if p.ValidateAuthCert() != nil {
		LogPrintln(p, "certificate file not found")
		return errors.New("certificate file not found")
	}

	return nil
}

func (p *Plugin) ValidateAuthCert() error {
	if p.AuthCert == "" {
		return nil
	}

	if _, err := os.Stat(p.AuthCert); os.IsNotExist(err) {
		return err
	}

	cert, err := os.ReadFile(p.AuthCert)
	if err != nil {
		return err
	}

	_, err = tls.X509KeyPair(cert, cert)
	if err != nil {
		return err
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

	if p.IsUploadFileRequired() {
		return p.AddFileUploadData()
	}

	bodyStr := p.RequestBody
	method := p.HttpMethod

	if bodyStr != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
		p.BodyIoReader = strings.NewReader(p.RequestBody)
	} else {
		p.BodyIoReader = nil
	}

	return nil
}

func (p *Plugin) AddFileUploadData() error {

	absoluteFilePath, err := GetAbsolutePath(p.UploadFile)
	if err != nil {
		return err
	}
	p.uploadFileAbsolutePath = absoluteFilePath

	if p.WrapAsMultipart {
		return p.AddFileUploadDataAsMultiPart()
	}

	return p.AddFileUploadDataWithoutMultiPart()
}

func (p *Plugin) AddFileUploadDataAsMultiPart() error {

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	file, err := os.Open(p.uploadFileAbsolutePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer func() {
		if file != nil {
			err := file.Close()
			if err != nil {
				return
			}
		}
	}()

	part, err := writer.CreateFormFile(p.MultiPartName, filepath.Base(p.uploadFileAbsolutePath))
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}
	defer func() {
		if writer == nil {
			return
		}
		err := writer.Close()
		if err != nil {
			return
		}
	}()

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error copying file content: %v", err)
	}

	p.BodyIoReader = &body
	p.ContentType = writer.FormDataContentType()

	return nil
}

func (p *Plugin) AddFileUploadDataWithoutMultiPart() error {

	file, err := os.Open(p.uploadFileAbsolutePath)
	if err != nil {
		LogPrintln(p, "error opening file: ", err.Error())
		return fmt.Errorf("error opening file: %v", err)
	}

	p.BodyIoReader = file
	p.ContentType = ApplicationOctetStream

	return nil
}

func (p *Plugin) IsUploadFileRequired() bool {
	return p.UploadFile != ""
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

	if headerStr == "" {
		return nil
	}

	headersList := strings.Split(headerStr, ",")

	for i, headerItem := range headersList {
		headerItem = strings.TrimSpace(headerItem)
		if i == 0 && headerItem == "" {
			LogPrintln(p, `if i == 0 && headerItem == ""`)
			return errors.New("malformed header: empty header")
		}

		kvPair := strings.SplitN(headerItem, ":", 2)
		if len(kvPair) != 2 {
			LogPrintln(p, `if len(kvPair) != 2`)
			return errors.New("malformed header: " + headerItem + " missing colon")
		}

		key := strings.TrimSpace(kvPair[0])
		if key == "" {
			LogPrintln(p, `if key == ""`)
			return errors.New("malformed header: " + headerItem + " empty header name")
		}

		value := strings.TrimSpace(kvPair[1])
		if value == "" {
			LogPrintln(p, `if value == ""`)
			return errors.New("malformed header: " + headerItem + " empty header value")
		}
	}

	return nil
}

func (p *PluginInputParams) EmitCommandLine() (string, string) {
	return EmitCommandLineForPluginStruct(*p)
}

//
//
