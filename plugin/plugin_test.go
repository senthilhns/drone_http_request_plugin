package plugin

import (
	"io/ioutil"
	"testing"
)

const (
	TestUrl                    = "https://httpbin.org" // A URL that supports different HTTP methods for testing.
	ContentTypeApplicationJson = "Content-Type:application/json"
)

func TestPluginHttpMethods(t *testing.T) {
	// Define the test cases
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

	// Loop through each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare arguments for the plugin
			args := Args{
				PluginInputParams: PluginInputParams{
					Url:         tc.url,
					HttpMethod:  tc.httpMethod,
					Timeout:     30,
					Headers:     tc.headers,
					RequestBody: tc.body,
				},
			}

			// Create a new Plugin object
			plugin := &Plugin{
				Args:                 args,
				PluginProcessingInfo: PluginProcessingInfo{},
			}

			// Run the plugin
			err := plugin.Run()
			if err != nil {
				t.Fatalf("Run() returned an error: %v", err)
			}

			// Check the response for requests that return a body (not HEAD or OPTIONS)
			if tc.httpMethod != "HEAD" && tc.httpMethod != "OPTIONS" {
				if plugin.HttpResponse.StatusCode != 200 {
					t.Errorf("Expected status 200, but got %d", plugin.HttpResponse.StatusCode)
				}

				body, err := ioutil.ReadAll(plugin.HttpResponse.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				_ = body
				// t.Logf("Response body for %s: %s", tc.httpMethod, string(body))
			}

			// Close the response body after the test
			plugin.HttpResponse.Body.Close()
		})
	}
}

func PrintTestLog(t *testing.T, msg string) {
	t.Logf("Test log: %s", msg)
}
