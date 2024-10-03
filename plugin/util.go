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
	"reflect"
	"strings"
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

func EmitCommandLineForPluginStruct(ifce interface{}) (string, string) {

	dockerImageName := "senthilhns/drone_http_request_plugin"

	shCli := EmitShellCommandLineForPluginStruct(ifce)
	dockerCli := EmitDockerShellCommandLineForPluginStruct(ifce, dockerImageName)

	return shCli, dockerCli
}

func EmitShellCommandLineForPluginStruct(ifce interface{}) string {

	shCli := func(envTagStr, valueStr string) string {
		s := fmt.Sprintf("%s='%s'", envTagStr, valueStr)
		return s
	}

	processSuffixArgs := func() []string {
		retStrList := []string{}

		retStrList = append(retStrList, shCli("TMP_PLUGIN_LOCAL_TESTING", "TRUE"))
		retStrList = append(retStrList, shCli("PLUGIN_IS_TESTING", "true"))
		retStrList = append(retStrList, `  go run ../main.go `)
		return retStrList
	}

	prefixArgs := func() []string {
		var prefixArgsList = []string{}
		return prefixArgsList
	}

	k := EmitShCliForPluginStruct(ifce, prefixArgs, shCli, processSuffixArgs)

	return k

}

func EmitDockerShellCommandLineForPluginStruct(ifce interface{}, dockerImageName string) string {

	prefixArgs := func() []string {
		var prefixArgsList = []string{}
		prefixArgsList = append(prefixArgsList, `docker run --rm`)
		return prefixArgsList
	}

	shCli := func(envTagStr, valueStr string) string {
		s := fmt.Sprintf("-e %s='%s'", envTagStr, valueStr)
		return s
	}

	processSuffixArgs := func() []string {
		retStrList := []string{}

		retStrList = append(retStrList, shCli("TMP_PLUGIN_LOCAL_TESTING", "TRUE"))
		retStrList = append(retStrList, shCli("PLUGIN_IS_TESTING", "true"))
		retStrList = append(retStrList, `  -w /drone/src  -v $(pwd):/drone/src `+dockerImageName)

		return retStrList
	}

	k := EmitShCliForPluginStruct(ifce, prefixArgs, shCli, processSuffixArgs)
	// fmt.Println(k)

	return k

}

func EmitShCliForPluginStruct(s interface{},
	processPrefixArgs func() []string,
	processTag func(envTagStr, valueStr string) string,
	processSuffixArgs func() []string) string {

	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	if v.Kind() != reflect.Struct {
		return ""
	}

	var envVars []string

	envVars = append(envVars, processPrefixArgs()...)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		envTag := field.Tag.Get("envconfig")
		if envTag == "" {
			continue
		}

		if value.IsZero() {
			continue
		}

		switch value.Kind() {
		case reflect.String:
			envVars = append(envVars, processTag(envTag, value.String()))
		case reflect.Int:
			envVars = append(envVars, processTag(envTag, fmt.Sprintf("%d", value.Int())))
		case reflect.Bool:
			envVars = append(envVars, processTag(envTag, fmt.Sprintf("%t", value.Bool())))
		}
	}

	suffixArgs := processSuffixArgs()
	envVars = append(envVars, suffixArgs...)
	envString := strings.Join(envVars, " \\\n")

	return fmt.Sprintf("%s \\\n ", envString)
}

func WriteCommandsListToFile(fileName string, commandsList []string) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString("#!/bin/bash\n\n")
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	for _, cmd := range commandsList {
		_, err := file.WriteString(cmd + "\n\n")
		if err != nil {
			fmt.Printf("Error writing command to file: %v\n", err)
			os.Exit(1)
		}
	}
}

func WriteEnvToFile(key string, value interface{}) error {

	outputFile, err := os.OpenFile(os.Getenv("DRONE_OUTPUT"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer outputFile.Close()

	valueStr := fmt.Sprintf("%v", value)

	_, err = fmt.Fprintf(outputFile, "%s=%s\n", key, valueStr)
	if err != nil {
		return fmt.Errorf("failed to write to env: %w", err)
	}

	return nil
}

const (
	Schema                 = "https://drone.github.io/drone-jira/card.json"
	StdOut                 = "/dev/stdout"
	ApplicationOctetStream = "application/octet-stream"
	ApplicationJson        = "application/json"
	ContentType            = "Content-Type"
)

//
//
