

$ boilr template download drone/boilr-plugin drone-plugin

$ boilr template use drone-plugin my-drone-plugin

Please choose a value for "DockerRepository" [default: "owner/name"]:
Please choose a value for "Description"
Please choose a value for "GoModule" [default: "github.com/owner/name":
Please choose a value for "GoModule


cd ../someprefix/go-convert
go build && ./go-convert jenkinsjson  ../drone_http_request_plugin/samples/http_request_trace_2.json

docker run --rm -e PLUGIN_PARAM1=foo -e PLUGIN_PARAM2=bar \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_BUILD_NUMBER=43 \
  -e DRONE_BUILD_STATUS=success \
  -w /drone/src \
  -v $(pwd):/drone/src \
  senthilhns/drone_http_request_plugin


docker run --rm \
-e PLUGIN_URL='http://127.0.0.1:36991' \
-e PLUGIN_HTTP_METHOD='POST' \
-e PLUGIN_QUIET='true' \
-e PLUGIN_UPLOAD_FILE='/home/hns/multipart-test1.txt' \
-e TMP_PLUGIN_LOCAL_TESTING='TRUE' \
-e PLUGIN_IS_TESTING='true' \
-e TMP_PLUGIN_LOCAL_TESTING='TRUE' \
-e PLUGIN_IS_TESTING='true' \
 go run ../main.go



docker run --rm -e PLUGIN_URL=foo \
  -w /drone/src \
  -v $(pwd):/drone/src \
  senthilhns/drone_http_request_plugin

#
#
