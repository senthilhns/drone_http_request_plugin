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

func TestPluginHttpMethods(t *testing.T) {
	tests := []struct {
		name       string
		httpMethod string
		url        string
		body       string
		headers    string
	}{
		{name: "GET request", httpMethod: "GET", url: TestUrl + "/get", body: "", headers: ContentTypeApplicationJson},
		{name: "POST request", httpMethod: "POST", url: TestUrl + "/post", body: `{"name":"drone"}`, headers: ContentTypeApplicationJson},
		{name: "PUT request", httpMethod: "PUT", url: TestUrl + "/put", body: `{"name":"drone"}`, headers: ContentTypeApplicationJson},
		{name: "DELETE request", httpMethod: "DELETE", url: TestUrl + "/delete", body: "", headers: ContentTypeApplicationJson},
		{name: "PATCH request", httpMethod: "PATCH", url: TestUrl + "/patch", body: `{"name":"drone"}`, headers: ContentTypeApplicationJson},
		{name: "HEAD request", httpMethod: "HEAD", url: TestUrl + "/get", body: "", headers: ContentTypeApplicationJson},
		{name: "OPTIONS request", httpMethod: "OPTIONS", url: TestUrl + "/get", body: "", headers: ContentTypeApplicationJson},
		{name: "MKCOL request", httpMethod: "MKCOL", url: TestUrl + "/mkcol", body: "", headers: ContentTypeApplicationJson},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := Args{
				PluginInputParams: PluginInputParams{
					Url:         tc.url,
					HttpMethod:  tc.httpMethod,
					Timeout:     30,
					Headers:     tc.headers,
					RequestBody: tc.body,
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

			if tc.httpMethod != "HEAD" && tc.httpMethod != "OPTIONS" {
				if plugin.httpResponse.StatusCode != 200 {
					t.Errorf("Expected status 200, but got %d", plugin.httpResponse.StatusCode)
				}

				body, err := ioutil.ReadAll(plugin.httpResponse.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				_ = body
				// t.Logf("Response body for %s: %s", tc.httpMethod, string(body))
			}

			plugin.httpResponse.Body.Close()
		})
	}
}

func TestBadHeaders(t *testing.T) {
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

	body, err := ioutil.ReadAll(plugin.httpResponse.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	t.Logf("Response body: %s", string(body))

}

func TestNegativeAuthBasic(t *testing.T) {
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
