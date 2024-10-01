package plugin

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func TestMKCOLWithLocalWebDAVServer(t *testing.T) {

	_, found := enableTests["TestMKCOLWithLocalWebDAVServer"]
	if !found {
		t.Skip("Skipping TestMKCOLWithLocalWebDAVServer test")
	}

	var ts *httptest.Server

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "MKCOL" {
			w.WriteHeader(http.StatusCreated)
			return
		}

		if r.URL.Path == "/quit" {
			w.WriteHeader(http.StatusOK)
			go ts.Close()
			return
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer ts.Close()

	randomDir := randomString(10)
	mkcolURL := fmt.Sprintf("%s/%s", ts.URL, randomDir)

	args := Args{
		PluginInputParams: PluginInputParams{
			Url:        mkcolURL,
			HttpMethod: "MKCOL",
			Quiet:      true,
		},
	}

	plugin := GetNewPlugin(args)

	thisTestName := "TestMKCOLWithLocalWebDAVServer"
	cli, dockerCli := plugin.EmitCommandLine()
	emittedCommands = append(emittedCommands, "# "+thisTestName+"\n"+cli)
	dockerCliCommands = append(dockerCliCommands, "# "+thisTestName+"\n"+dockerCli)

	err := plugin.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("MKCOL", mkcolURL, nil)
	if err != nil {
		t.Fatalf("Error creating MKCOL request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error making MKCOL request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, but got %d", resp.StatusCode)
	}
	t.Logf("MKCOL request succeeded for directory: %s", randomDir)

	quitURL := fmt.Sprintf("%s/quit", ts.URL)
	req, err = http.NewRequest("GET", quitURL, nil)
	if err != nil {
		t.Fatalf("Error creating shutdown request: %v", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Error sending shutdown request: %v", err)
	}
	defer resp.Body.Close()

	plugin.DeInit()
}

//
//
