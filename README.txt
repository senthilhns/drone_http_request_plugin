docker run --rm -e PLUGIN_PARAM1=foo -e PLUGIN_PARAM2=bar \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_BUILD_NUMBER=43 \
  -e DRONE_BUILD_STATUS=success \
  -w /drone/src \
  -v $(pwd):/drone/src \
  senthilhns/drone_http_request_plugin

docker run --rm -e PLUGIN_URL=foo \
  -w /drone/src \
  -v $(pwd):/drone/src \
  senthilhns/drone_http_request_plugin
