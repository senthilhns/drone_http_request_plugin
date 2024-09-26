package plugin

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func startWebDAVServer(t *testing.T) *exec.Cmd {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	serverPath := filepath.Join(workingDir, "../test-resources/webdav_server.go")
	cmd := exec.Command("go", "run", serverPath)

	err = cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start WebDAV server: %v", err)
	}

	time.Sleep(15 * time.Second)

	return cmd
}

func TestMKCOLWithLocalWebDAVServer(t *testing.T) {

	_, found := enableTests["TestMKCOLWithLocalWebDAVServer"]
	if !found {
		t.Skip("Skipping TestMKCOLWithLocalWebDAVServer test")
	}

	fmt.Println("Starting WEB server test waits for 10 seconds for webdav server to start")
	cmd := startWebDAVServer(t)

	randomDir := randomString(10)
	mkcolURL := fmt.Sprintf("http://localhost:8080/%s", randomDir)

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

	quitURL := "http://localhost:8080/quit"
	req, err = http.NewRequest("GET", quitURL, nil)
	if err != nil {
		t.Fatalf("Error creating shutdown request: %v", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Error sending shutdown request: %v", err)
	}
	defer resp.Body.Close()

	err = cmd.Wait()
	if err != nil {
		t.Fatalf("Error waiting for server shutdown: %v", err)
	}
	t.Log("Server shutdown successfully")
}

//
//
