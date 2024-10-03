package plugin

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	TestUrl                    = "https://httpbin.org"
	ContentTypeApplicationJson = "Content-Type:application/json"
)

var emittedCommands []string
var dockerCliCommands []string

var enableTests = map[string]bool{
	"TestGetRequest":                      true,
	"TestGetRequestWithValidResponseBody": true,
	"TestPostRequest":                     true,
	"TestPutRequest":                      true,
	"TestDeleteRequest":                   true,
	"TestPatchRequest":                    true,
	"TestHeadRequest":                     true,
	"TestOptionsRequest":                  true,
	"TestMkcolRequest":                    true,
	"TestMKCOLWithLocalWebDAVServer":      true,

	"TestGetRequestAndWriteToFile":         true,
	"TestGetRequestWithResponseLogging":    true,
	"TestGetRequestWithoutResponseLogging": true,
	"TestGetRequestWithQuietMode":          true,
	"TestDirectFileUpload":                 true,
	"TestMultipartFileUpload":              true,

	"TestGetRequestUsingProxyWithPlugin": true,

	"TestNegativeAuthBasic": true,
	"TestPositiveAuthBasic": true,

	"TestGetRequestWithAcceptType":          true,
	"TestGetRequestWithIncorrectAcceptType": true,

	//"TestSSlRequiredNoClientCertNoProxy": true,
	//"TestSSlRequiredClientCertNoProxy":   true,
	//"TestSslSkippingNoClientCertNoProxy": true,
	//"TestSslSkippingClientCertNoProxy":   true,
	//"TestSSlRequiredNoClientCertProxyEnabled": true,
	//"TestSSlRequiredClientCertProxyEnabled":   true,
	//"TestSslSkippingClientCertProxyEnabled":   true,

	//"TestSslSkippingNoClientCertProxyEnabled": true,
}

const IsEmitCli = true

//const IsEmitCli = false

func TestMain(m *testing.M) {

	exitCode := m.Run()

	if !IsEmitCli {
		os.Exit(exitCode)
	}

	WriteCommandsListToFile("/tmp/run_plugin_cli.sh", emittedCommands)
	WriteCommandsListToFile("/tmp/run_docker_cli.sh", dockerCliCommands)

	os.Exit(exitCode)
}

func TestGetRequest(t *testing.T) {
	thisTestName := "TestGetRequest"
	_, found := enableTests[thisTestName]
	if !found {
		t.Skip("Skipping " + thisTestName + " test")
	}

	cli := runPluginTest(t, "GET", TestUrl+"/get", "", ContentTypeApplicationJson)

	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
}

func TestGetRequestWithValidResponseBody(t *testing.T) {

	thisTestName := "TestGetRequestWithValidResponseBody"

	_, found := enableTests[thisTestName]
	if !found {
		t.Skip("Skipping " + thisTestName + " test")
	}

	expectedResponseBody := `"Content-Type": "application/json",`

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:               TestUrl + "/get",
			HttpMethod:        "GET",
			Timeout:           30,
			Headers:           ContentTypeApplicationJson,
			ValidResponseBody: expectedResponseBody,
		},
	}

	plugin := GetNewPlugin(args)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	defer func() {
		plugin.DeInit()
	}()

	if plugin.httpResponse.StatusCode != 200 {
		t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
	}

	if !strings.Contains(plugin.ResponseContent, expectedResponseBody) {
		t.Errorf("Expected response body to contain %q, but it did not. Response body: %s", expectedResponseBody, plugin.ResponseContent)
	} else {
		t.Logf("Test passed. Response body contains the expected substring: %s", expectedResponseBody)
	}
}

func TestPostRequest(t *testing.T) {

	thisTestName := "TestPostRequest"

	_, found := enableTests[thisTestName]
	if !found {
		t.Skip("Skipping TestPostRequest test")
	}

	cli := runPluginTest(t, "POST", TestUrl+"/post", `{"name":"drone"}`, ContentTypeApplicationJson)
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)

}

func TestPutRequest(t *testing.T) {

	thisTestName := "TestPutRequest"

	_, found := enableTests["TestPutRequest"]
	if !found {
		t.Skip("Skipping TestPutRequest test")
	}

	cli := runPluginTest(t, "PUT", TestUrl+"/put", `{"name":"drone"}`, ContentTypeApplicationJson)
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
}

func TestDeleteRequest(t *testing.T) {

	thisTestName := "TestDeleteRequest"

	_, found := enableTests["TestDeleteRequest"]
	if !found {
		t.Skip("Skipping TestDeleteRequest test")
	}

	cli := runPluginTest(t, "DELETE", TestUrl+"/delete", "", ContentTypeApplicationJson)
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
}

func TestPatchRequest(t *testing.T) {

	thisTestName := "TestPatchRequest"
	_, found := enableTests["TestPatchRequest"]
	if !found {
		t.Skip("Skipping TestPatchRequest test")
	}

	cli := runPluginTest(t, "PATCH", TestUrl+"/patch", `{"name":"drone"}`, ContentTypeApplicationJson)
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
}

func TestHeadRequest(t *testing.T) {

	thisTestName := "TestHeadRequest"
	_, found := enableTests["TestHeadRequest"]
	if !found {
		t.Skip("Skipping TestHeadRequest test")
	}

	cli := runPluginTest(t, "HEAD", TestUrl+"/get", "", ContentTypeApplicationJson)
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
}

func TestOptionsRequest(t *testing.T) {

	thisTestName := "TestOptionsRequest"
	_, found := enableTests["TestOptionsRequest"]
	if !found {
		t.Skip("Skipping TestOptionsRequest test")
	}

	cli := runPluginTest(t, "OPTIONS", TestUrl+"/get", "", ContentTypeApplicationJson)
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
}

func TestBadHeaders(t *testing.T) {

	_, found := enableTests["TestBadHeaders"]
	if !found {
		t.Skip("Skipping TestPluginHttpMethods test")
	}

	tests := []struct {
		name    string
		headers string
		wantErr bool
	}{
		{name: "Missing colon", headers: "Authorization Bearer token", wantErr: true},
		{name: "Empty header name", headers: ": value", wantErr: true},
		{name: "Empty header value", headers: "Content-Type:", wantErr: true},
		{name: "Valid header", headers: "Content-Type: application/json", wantErr: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := Args{
				PluginInputParams: PluginInputParams{
					Url:        TestUrl + "/get",
					HttpMethod: "GET",
					Headers:    tc.headers,
					Timeout:    30,
				},
			}

			plugin := &Plugin{
				Args:                 args,
				PluginProcessingInfo: PluginProcessingInfo{},
			}

			err := plugin.ValidateHeader(tc.headers)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected an error for headers: %q, but got none", tc.headers)
				} else {
					t.Logf("Got expected error for bad headers: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error for headers: %q, but got: %v", tc.headers, err)
				}
			}
		})
	}
}

func TestPositiveAuthBasic(t *testing.T) {

	_, found := enableTests["TestPositiveAuthBasic"]
	if !found {
		t.Skip("Skipping TestPositiveAuthBasic test")
	}

	username := "user"
	password := "pass"

	expectedAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        "https://httpbin.org/basic-auth/user/pass",
			HttpMethod: "GET",
			AuthBasic:  username + ":" + password,
			Timeout:    30,
			Headers:    ContentTypeApplicationJson,
		},
	}

	plugin := &Plugin{
		Args:                 args,
		PluginProcessingInfo: PluginProcessingInfo{},
	}

	thisTestName := "TestPositiveAuthBasic"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}
	defer func() {
		plugin.DeInit()
	}()

	authHeader := plugin.HttpReq.Header.Get("Authorization")
	if authHeader != expectedAuthHeader {
		t.Errorf("Expected Authorization header to be %q, but got %q", expectedAuthHeader, authHeader)
	}

	if plugin.httpResponse.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
	}

	body, err := ioutil.ReadAll(plugin.httpResponse.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	t.Logf("Response body: %s", string(body))

}

func TestNegativeAuthBasic(t *testing.T) {

	_, found := enableTests["TestNegativeAuthBasic"]
	if !found {
		t.Skip("Skipping TestNegativeAuthBasic test")
	}

	username := "user"
	password := "pass"

	expectedAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        "https://httpbin.org/basic-auth/user/pass",
			HttpMethod: "GET",
			AuthBasic:  username + ":" + password,
			Timeout:    30,
		},
	}

	plugin := &Plugin{
		Args:                 args,
		PluginProcessingInfo: PluginProcessingInfo{},
	}

	thisTestName := "TestNegativeAuthBasic"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}
	defer func() {
		plugin.DeInit()
	}()

	authHeader := plugin.HttpReq.Header.Get("Authorization")
	if authHeader != expectedAuthHeader {
		t.Errorf("Expected Authorization header to be %q, but got %q", expectedAuthHeader, authHeader)
	}

	if plugin.httpResponse.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
	}

}

func randomFileName() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("output_%d.txt", rand.Intn(100000))
}

func TestGetRequestAndWriteToFile(t *testing.T) {

	_, found := enableTests["TestGetRequestAndWriteToFile"]
	if !found {
		t.Skip("Skipping TestGetRequestAndWriteToFile test")
	}

	outputFile := "/tmp/" + randomFileName()

	//defer os.Remove(outputFile)

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        "https://httpbin.org/get",
			HttpMethod: "GET",
			Timeout:    30,
			Headers:    ContentTypeApplicationJson,
			OutputFile: outputFile,
		},
	}

	plugin := GetNewPlugin(args)

	thisTestName := "TestGetRequestAndWriteToFile"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Expected output file %s to be created, but it does not exist", outputFile)
	}

	content, err := ioutil.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read the output file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("Output file %s is empty, expected response body to be written", outputFile)
	}

	t.Logf("Test passed. Response written to file: %s", outputFile)
	t.Logf("Response content: %s", string(content))
}

func TestGetRequestWithResponseLogging(t *testing.T) {
	_, found := enableTests["TestGetRequestWithResponseLogging"]
	if !found {
		t.Skip("Skipping TestGetRequestWithResponseLogging test")
	}
	CheckForResponseLogging(t, true)
}

func TestGetRequestWithoutResponseLogging(t *testing.T) {
	_, found := enableTests["TestGetRequestWithoutResponseLogging"]
	if !found {
		t.Skip("Skipping TestGetRequestWithoutResponseLogging test")
	}
	CheckForResponseLogging(t, false)
}

func CheckForResponseLogging(t *testing.T, isLogResponse bool) {

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(ioutil.Discard)

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:         TestUrl + "/get",
			HttpMethod:  "GET",
			Timeout:     30,
			Headers:     ContentTypeApplicationJson,
			LogResponse: isLogResponse,
		},
	}

	plugin := GetNewPlugin(args)
	if plugin == nil {
		if isLogResponse {
			plugin.LogResponse = true
		} else {
			plugin.LogResponse = false
		}
	}

	if !isLogResponse {
		plugin.IsSuppressLogs = true
	}

	thisTestName := "TestGetRequestWithResponseLogging"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	defer func() {
		plugin.DeInit()
	}()

	if plugin.httpResponse.StatusCode != 200 {
		t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
	}

	if isLogResponse {
		if !strings.Contains(logBuffer.String(), plugin.ResponseContent) {
			t.Errorf("Expected response content to be logged, but it was not. Logged content")
		} else {
			t.Logf("Test passed. Response content was logged.")
		}
	} else {
		if strings.Contains(logBuffer.String(), plugin.ResponseContent) {
			t.Errorf("Expected response content not to be logged, but it was. "+
				"Logged content logBuffer \n %s \n ResponseContent \n %s \n",
				logBuffer.String(), plugin.ResponseContent)
		} else {
			t.Logf("Test passed. Response content was not logged.")
		}
	}
}

func TestPluginWithCustomSslCert(t *testing.T) {

	thisTestName := "TestPluginWithCustomSslCert"

	_, found := enableTests["TestPluginWithCustomSslCert"]
	if !found {
		t.Skip("Skipping TestPluginWithCustomSslCert test")
	}

	const certName = "./bogus_private.key.pem"

	if _, err := os.Stat(certName); os.IsNotExist(err) {
		t.Fatalf("Certificate file not found: %v", err)
	}

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        "https://httpbin.org/get",
			HttpMethod: "GET",
			Timeout:    30,
			Headers:    ContentTypeApplicationJson,
			AuthCert:   certName,
			IgnoreSsl:  false,
		},
	}

	plugin := GetNewPlugin(args)

	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	plugin.httpResponse = &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(`{"status":"ok"}`)),
	}

	err = plugin.StoreHttpResponseResults()
	if err != nil {
		t.Fatalf("Failed to store HTTP response results: %v", err)
	}

	if plugin.ResponseStatus != 200 {
		t.Errorf("Expected response status 200, got %d", plugin.ResponseStatus)
	}

	if plugin.ResponseContent != `{"status":"ok"}` {
		t.Errorf("Expected response content to be '%s', got '%s'", `{"status":"ok"}`, plugin.ResponseContent)
	}
}

func runPluginTest(t *testing.T, method, url, body, headers string) string {
	args := Args{
		PluginInputParams: PluginInputParams{
			Url:         url,
			HttpMethod:  method,
			Timeout:     30,
			Headers:     headers,
			RequestBody: body,
		},
	}

	plugin := &Plugin{
		Args:                 args,
		PluginProcessingInfo: PluginProcessingInfo{},
	}

	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+method+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+method+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		LogPrintln(plugin, "Error: ", err.Error())
		t.Fatalf("Run() returned an error: %v", err)
	}

	defer func() {
		plugin.DeInit()
	}()

	if method != "HEAD" && method != "OPTIONS" {
		if plugin.httpResponse.StatusCode != 200 {
			t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
		}

		body, err := io.ReadAll(plugin.httpResponse.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		_ = body
	}

	plugin.httpResponse.Body.Close()
	return cli
}

func TestGetRequestWithQuietMode(t *testing.T) {
	_, found := enableTests["TestGetRequestWithQuietMode"]
	if !found {
		t.Skip("Skipping TestGetRequestWithQuietMode test")
	}

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(io.Discard)

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:         TestUrl + "/get",
			HttpMethod:  "GET",
			Timeout:     30,
			Headers:     ContentTypeApplicationJson,
			LogResponse: true,
			Quiet:       true,
		},
	}

	plugin := GetNewPlugin(args)

	thisTestName := "TestGetRequestWithQuietMode"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	defer func() {
		plugin.DeInit()
	}()

	if plugin.httpResponse.StatusCode != 200 {
		t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
	}

	if logBuffer.Len() > 0 {
		t.Errorf("Expected no logs in Quiet mode, but some logs were written: %s", logBuffer.String())
	} else {
		t.Logf("Test passed. No logs were written in Quiet mode.")
	}
}

func TestMultipartFileUpload(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, "multipart-test.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(filePath)
	file.WriteString("test content")
	file.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Expected multipart/form-data, got %s", r.Header.Get("Content-Type"))
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Errorf("Error parsing multipart form: %v", err)
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("Expected file, but got error: %v", err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil || string(content) != "test content" {
			t.Errorf("Expected 'test content', but got %s", string(content))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:             ts.URL,
			HttpMethod:      "POST",
			WrapAsMultipart: true,
			UploadFile:      filePath,
			MultiPartName:   "file",
			Quiet:           true,
		},
	}

	plugin := GetNewPlugin(args)

	err = plugin.Run()

	thisTestName := "TestMultipartFileUpload"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	defer func() {
		plugin.DeInit()
	}()

}

func TestDirectFileUpload(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, "multipart-test1.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file.WriteString("test content")
	defer file.Close()
	defer os.Remove(filePath)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Errorf("Expected application/octet-stream, got %s", r.Header.Get("Content-Type"))
		}
		content, err := io.ReadAll(r.Body)
		if err != nil || string(content) != "test content" {
			t.Errorf("Expected 'test content', but got %s", string(content))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:             ts.URL,
			HttpMethod:      "POST",
			WrapAsMultipart: false,
			UploadFile:      filePath,
			Quiet:           true,
		},
	}

	plugin := GetNewPlugin(args)

	thisTestName := "TestDirectFileUpload"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err = plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	defer func() {
		plugin.DeInit()
	}()

}

// this should pass to check whether proxy is fine
func TestGetRequestUsingProxyWithoutPlugin(t *testing.T) {
	proxyURL, _ := url.Parse("http://localhost:8888")

	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		t.Fatalf("Failed to send GET request through proxy: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response from httpbin: %s\n", body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", resp.StatusCode)
	}
}

/*
check these first
docker run -d --name='tinyproxy' -p 8888:8888 dannydirect/tinyproxy:latest ANY
curl -x http://localhost:8888 https://httpbin.org/ip -vv
*/

var ProxyURL = "http://localhost:8888"

func TestGetRequestUsingProxyWithPlugin(t *testing.T) {
	thisTestName := "TestGetRequestUsingProxyWithPlugin"
	_, found := enableTests[thisTestName]
	if !found {
		t.Skip("Skipping " + thisTestName + " test")
	}

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        TestUrl + "/get",
			HttpMethod: "GET",
			Headers:    ContentTypeApplicationJson,
			Proxy:      ProxyURL,
			Timeout:    30,
			IgnoreSsl:  true,
		},
	}

	plugin := GetNewPlugin(args)
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	defer func() {
		plugin.DeInit()
	}()

	if plugin.httpResponse.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
	}

	body, err := io.ReadAll(plugin.httpResponse.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("Response from httpbin: %s", body)
}

func TestGetRequestWithAcceptType(t *testing.T) {
	thisTestName := "TestGetRequestWithAcceptType"
	expectedAcceptType := "application/json"

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        "https://httpbin.org/get",
			HttpMethod: "GET",
			Timeout:    30,
			AcceptType: expectedAcceptType,
			Headers:    "Content-Type:application/json",
		},
	}

	plugin := GetNewPlugin(args)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptHeader := r.Header.Get("Accept")
		if acceptHeader != expectedAcceptType {
			t.Errorf("Expected Accept header to be %s, but got %s", expectedAcceptType, acceptHeader)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	plugin.Args.PluginInputParams.Url = ts.URL

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	defer plugin.DeInit()

	if plugin.httpResponse.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
	}

	t.Logf("Test passed. Accept header correctly set to %s.", expectedAcceptType)
}

func TestGetRequestWithIncorrectAcceptType(t *testing.T) {
	thisTestName := "TestGetRequestWithIncorrectAcceptType"
	expectedAcceptType := "application/xml"

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        "https://httpbin.org/get",
			HttpMethod: "GET",
			Timeout:    30,
			AcceptType: expectedAcceptType,
			Headers:    "Content-Type:application/json",
		},
	}

	plugin := GetNewPlugin(args)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptHeader := r.Header.Get("Accept")
		if acceptHeader != "application/json" {
			http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	plugin.Args.PluginInputParams.Url = ts.URL

	err := plugin.Run()

	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	defer plugin.DeInit()

	if err == nil {
		t.Fatalf("Expected an error due to incorrect AcceptType, but Run() did not return an error")
	}

	if plugin.httpResponse.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("Expected status 415 Unsupported Media Type, but got %d", plugin.httpResponse.StatusCode)
	}

	t.Logf("Test passed. Incorrect Accept header was correctly rejected by the server.")
}
