// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func writeCard(path, schema string, card interface{}) {
	data, _ := json.Marshal(map[string]interface{}{
		"schema": schema,
		"data":   card,
	})
	switch {
	case path == "/dev/stdout":
		writeCardTo(os.Stdout, data)
	case path == "/dev/stderr":
		writeCardTo(os.Stderr, data)
	case path != "":
		ioutil.WriteFile(path, data, 0644)
	}
}

func writeCardTo(out io.Writer, data []byte) {

	if os.Getenv("TMP_PLUGIN_LOCAL_TESTING") == "TRUE" {
		LogPrintln("writeCardTo TMP_PLUGIN_LOCAL_TESTING is TRUE, skipping writing card")
		return
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	io.WriteString(out, "\u001B]1338;")
	io.WriteString(out, encoded)
	io.WriteString(out, "\u001B]0m")
	io.WriteString(out, "\n")
}

func LogPrintln(args ...interface{}) {
	log.Println(append([]interface{}{"Plugin Info:"}, args...)...)
}

func LogPrintf(format string, args ...interface{}) {
	log.Printf("Plugin Info: "+format, args...)
}

//
//
