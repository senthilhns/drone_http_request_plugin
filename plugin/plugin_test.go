package plugin

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	TestUrl                    = "https://httpbin.org"
	ContentTypeApplicationJson = "Content-Type:application/json"
)

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

	//"TestSSlRequiredNoClientCertNoProxy": true,
	//"TestSSlRequiredClientCertNoProxy":   true,
	//"TestSslSkippingNoClientCertNoProxy": true,
	//"TestSslSkippingClientCertNoProxy":   true,
	//"TestSSlRequiredNoClientCertProxyEnabled": true,
	//"TestSSlRequiredClientCertProxyEnabled":   true,
	//"TestSslSkippingClientCertProxyEnabled":   true,

	// "TestSslSkippingNoClientCertProxyEnabled": true,
}

func TestGetRequest(t *testing.T) {
	_, found := enableTests["TestGetRequest"]
	if !found {
		t.Skip("Skipping TestGetRequest test")
	}

	runPluginTest(t, "GET", TestUrl+"/get", "", ContentTypeApplicationJson)
}

func TestGetRequestWithValidResponseBody(t *testing.T) {

	_, found := enableTests["TestGetRequestWithValidResponseBody"]
	if !found {
		t.Skip("Skipping TestGetRequestWithValidResponseBody test")
	}

	expectedResponseBody := `"Content-Type": "application/json",`

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:               TestUrl + "/get", // A simple URL to perform a GET request
			HttpMethod:        "GET",
			Timeout:           30,
			Headers:           ContentTypeApplicationJson,
			ValidResponseBody: expectedResponseBody, // Set the expected substring
		},
	}

	plugin := GetNewPlugin(args)

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

	if !strings.Contains(plugin.ResponseContent, expectedResponseBody) {
		t.Errorf("Expected response body to contain %q, but it did not. Response body: %s", expectedResponseBody, plugin.ResponseContent)
	} else {
		t.Logf("Test passed. Response body contains the expected substring: %s", expectedResponseBody)
	}
}

func TestPostRequest(t *testing.T) {
	_, found := enableTests["TestPostRequest"]
	if !found {
		t.Skip("Skipping TestPostRequest test")
	}

	runPluginTest(t, "POST", TestUrl+"/post", `{"name":"drone"}`, ContentTypeApplicationJson)
}

func TestPutRequest(t *testing.T) {
	_, found := enableTests["TestPutRequest"]
	if !found {
		t.Skip("Skipping TestPutRequest test")
	}

	runPluginTest(t, "PUT", TestUrl+"/put", `{"name":"drone"}`, ContentTypeApplicationJson)
}

func TestDeleteRequest(t *testing.T) {
	_, found := enableTests["TestDeleteRequest"]
	if !found {
		t.Skip("Skipping TestDeleteRequest test")
	}

	runPluginTest(t, "DELETE", TestUrl+"/delete", "", ContentTypeApplicationJson)
}

func TestPatchRequest(t *testing.T) {
	_, found := enableTests["TestPatchRequest"]
	if !found {
		t.Skip("Skipping TestPatchRequest test")
	}

	runPluginTest(t, "PATCH", TestUrl+"/patch", `{"name":"drone"}`, ContentTypeApplicationJson)
}

func TestHeadRequest(t *testing.T) {
	_, found := enableTests["TestHeadRequest"]
	if !found {
		t.Skip("Skipping TestHeadRequest test")
	}

	runPluginTest(t, "HEAD", TestUrl+"/get", "", ContentTypeApplicationJson)
}

func TestOptionsRequest(t *testing.T) {
	_, found := enableTests["TestOptionsRequest"]
	if !found {
		t.Skip("Skipping TestOptionsRequest test")
	}

	runPluginTest(t, "OPTIONS", TestUrl+"/get", "", ContentTypeApplicationJson)
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
			//AuthBasic:  username + ":" + password,
			Timeout: 30,
			Headers: ContentTypeApplicationJson,
		},
	}

	plugin := &Plugin{
		Args:                 args,
		PluginProcessingInfo: PluginProcessingInfo{},
	}

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
			Url:        "https://httpbin.org/get", // A simple URL to perform a GET request
			HttpMethod: "GET",
			Timeout:    30,
			Headers:    ContentTypeApplicationJson,
			OutputFile: outputFile, // Set the random output file
		},
	}

	plugin := GetNewPlugin(args)

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
			Url:        "https://httpbin.org/get", // A simple URL to perform a GET request
			HttpMethod: "GET",
			Timeout:    30,
			Headers:    ContentTypeApplicationJson,
			AuthCert:   certName,
			IgnoreSsl:  false,
		},
	}

	plugin := GetNewPlugin(args)

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

//
//

func runPluginTest(t *testing.T, method, url, body, headers string) {
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
		_ = body // You can log the body if necessary
	}

	plugin.httpResponse.Body.Close()
}

func TestGetRequestWithQuietMode(t *testing.T) {
	_, found := enableTests["TestGetRequestWithQuietMode"]
	if !found {
		t.Skip("Skipping TestGetRequestWithQuietMode test")
	}

	// Prepare a buffer to capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(ioutil.Discard) // Restore the default behavior after the test

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
