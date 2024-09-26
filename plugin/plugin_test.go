package plugin

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	TestUrl                    = "https://httpbin.org"
	ContentTypeApplicationJson = "Content-Type:application/json"
)

var enableTests = map[string]bool{
	"TestGetRequest":     true,
	"TestPostRequest":    true,
	"TestPutRequest":     true,
	"TestDeleteRequest":  true,
	"TestPatchRequest":   true,
	"TestHeadRequest":    true,
	"TestOptionsRequest": true,
	"TestMkcolRequest":   true,
}

func TestGetRequest(t *testing.T) {
	_, found := enableTests["TestGetRequest"]
	if !found {
		t.Skip("Skipping TestGetRequest test")
	}

	runPluginTest(t, "GET", TestUrl+"/get", "", ContentTypeApplicationJson)
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

func TestMkcolRequest(t *testing.T) {
	_, found := enableTests["TestMkcolRequest"]
	if !found {
		t.Skip("Skipping TestMkcolRequest test")
	}

	runPluginTest(t, "MKCOL", TestUrl+"/mkcol", "", ContentTypeApplicationJson)
}

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
		t.Fatalf("Run() returned an error: %v", err)
	}

	defer func() {
		plugin.DeInit()
	}()

	if method != "HEAD" && method != "OPTIONS" {
		if plugin.httpResponse.StatusCode != 200 {
			t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
		}

		body, err := ioutil.ReadAll(plugin.httpResponse.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		_ = body // You can log the body if necessary
	}

	plugin.httpResponse.Body.Close()
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

//
//
