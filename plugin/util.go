// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func writeCard(p *Plugin, path, schema string, card interface{}) {
	data, _ := json.Marshal(map[string]interface{}{
		"schema": schema,
		"data":   card,
	})
	switch {
	case path == "/dev/stdout":
		writeCardTo(p, os.Stdout, data)
	case path == "/dev/stderr":
		writeCardTo(p, os.Stderr, data)
	case path != "":
		ioutil.WriteFile(path, data, 0644)
	}
}

func writeCardTo(p *Plugin, out io.Writer, data []byte) {

	if os.Getenv("TMP_PLUGIN_LOCAL_TESTING") == "TRUE" {
		LogPrintln(p, "writeCardTo TMP_PLUGIN_LOCAL_TESTING is TRUE, skipping writing card")
		return
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	io.WriteString(out, "\u001B]1338;")
	io.WriteString(out, encoded)
	io.WriteString(out, "\u001B]0m")
	io.WriteString(out, "\n")
}

func LogPrintln(p *Plugin, args ...interface{}) {
	if p != nil {
		if p.Quiet {
			return
		}
	}

	log.Println(append([]interface{}{"Plugin Info:"}, args...)...)
}

func LogPrintf(p *Plugin, format string, args ...interface{}) {
	if p != nil {
		if p.Quiet {
			return
		}
	}

	log.Printf("Plugin Info: "+format, args...)
}

func GetAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %v", err)
	}

	return absPath, nil
}

//
//
