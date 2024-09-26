// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"io/ioutil"
	"testing"
)

const (
	TestUrl                    = "https://httpbin.org/get"
	ContentTypeApplicationJson = "Content-Type:application/json"
)

func TestPlugin(t *testing.T) {
	t.Skip()
}

func TestGetRequest(t *testing.T) {
	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        TestUrl,
			HttpMethod: "GET",
			Timeout:    30,
			Headers:    ContentTypeApplicationJson,
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

	if plugin.HttpResponse.StatusCode != 200 {
		t.Errorf("Expected status 200, but got %d", plugin.HttpResponse.StatusCode)
	}

	body, err := ioutil.ReadAll(plugin.HttpResponse.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("Response body: %s", string(body))

	// Close the response body after the test
	plugin.HttpResponse.Body.Close()
}
