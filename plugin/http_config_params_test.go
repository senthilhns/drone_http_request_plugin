package plugin

//import (
//	"os"
//	"testing"
//	"time"
//)
//
//const delayTime = 2
//
//// const ProxyUrl = "http://myproxy.com:8080"
//
//const ProxyUrl = "http://47.89.184.18:3128"
//
//func TestSSlRequiredNoClientCertNoProxy(t *testing.T) {
//
//	_, found := enableTests["TestSSlRequiredNoClientCertNoProxy"]
//	if !found {
//		t.Skip("Skipping TestSSlRequiredNoClientCertNoProxy test")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "https://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   false,
//			SslCertPath: "",
//			Proxy:       "",
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
//func TestSSlRequiredClientCertNoProxy(t *testing.T) {
//
//	_, found := enableTests["TestSSlRequiredClientCertNoProxy"]
//	if !found {
//		t.Skip("Skipping TestSSlRequiredClientCertNoProxy test")
//	}
//
//	clientTestPath := os.Getenv("CLIENT_TEST_PATH")
//	if clientTestPath == "" {
//		t.Fatal("CLIENT_TEST_PATH environment variable is not set")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "https://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   false,
//			SslCertPath: clientTestPath,
//			Proxy:       "",
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
//func TestSslSkippingNoClientCertNoProxy(t *testing.T) {
//
//	_, found := enableTests["TestSslSkippingNoClientCertNoProxy"]
//	if !found {
//		t.Skip("Skipping TestSslSkippingNoClientCertNoProxy test")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "https://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   true,
//			SslCertPath: "",
//			Proxy:       "",
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
//func TestSslSkippingClientCertNoProxy(t *testing.T) {
//
//	_, found := enableTests["TestSslSkippingClientCertNoProxy"]
//	if !found {
//		t.Skip("Skipping TestSslSkippingClientCertNoProxy test")
//	}
//
//	clientTestPath := os.Getenv("CLIENT_TEST_PATH")
//	if clientTestPath == "" {
//		t.Fatal("CLIENT_TEST_PATH environment variable is not set")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "https://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   true,
//			SslCertPath: clientTestPath,
//			Proxy:       "",
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
//func TestSslSkippingClientCertProxyEnabled(t *testing.T) {
//
//	_, found := enableTests["TestSslSkippingClientCertProxyEnabled"]
//	if !found {
//		t.Skip("Skipping TestSslSkippingClientCertProxyEnabled test")
//	}
//
//	clientTestPath := os.Getenv("CLIENT_TEST_PATH")
//	if clientTestPath == "" {
//		t.Fatal("CLIENT_TEST_PATH environment variable is not set")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "https://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   true,
//			SslCertPath: clientTestPath,
//			Proxy:       ProxyUrl,
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
//func TestSSlRequiredNoClientCertProxyEnabled(t *testing.T) {
//
//	_, found := enableTests["TestSSlRequiredNoClientCertProxyEnabled"]
//	if !found {
//		t.Skip("Skipping TestSSlRequiredNoClientCertProxyEnabled test")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "https://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   false,
//			SslCertPath: "",
//			Proxy:       ProxyUrl,
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
//func TestSslSkippingNoClientCertProxyEnabled(t *testing.T) {
//
//	_, found := enableTests["TestSslSkippingNoClientCertProxyEnabled"]
//	if !found {
//		t.Skip("Skipping TestSslSkippingNoClientCertProxyEnabled test")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "http://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   true,
//			SslCertPath: "",
//			Proxy:       ProxyUrl,
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
//func TestSSlRequiredClientCertProxyEnabled(t *testing.T) {
//
//	_, found := enableTests["TestSSlRequiredClientCertProxyEnabled"]
//	if !found {
//		t.Skip("Skipping TestSSlRequiredClientCertProxyEnabled test")
//	}
//
//	clientTestPath := os.Getenv("CLIENT_TEST_PATH")
//	if clientTestPath == "" {
//		t.Fatal("CLIENT_TEST_PATH environment variable is not set")
//	}
//
//	args := Args{
//		PluginInputParams: PluginInputParams{
//			Url:         "https://example.com",
//			HttpMethod:  "GET",
//			Timeout:     30,
//			Headers:     "Content-Type: application/json",
//			RequestBody: `{"key":"value"}`,
//			IgnoreSsl:   false,
//			SslCertPath: clientTestPath,
//			Proxy:       ProxyUrl,
//		},
//	}
//
//	plugin := &Plugin{
//		Args:                 args,
//		PluginProcessingInfo: PluginProcessingInfo{},
//	}
//
//	err := plugin.Run()
//	if err != nil {
//		t.Fatalf("Run() returned an error: %v", err)
//	}
//	time.Sleep(delayTime * time.Second)
//}
//
////
////
