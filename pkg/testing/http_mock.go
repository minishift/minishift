/*
Copyright (C) 2016 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	"bufio"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

var DefaultRoundTripper http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func ResetDefaultRoundTripper() {
	http.DefaultClient.Transport = DefaultRoundTripper
}

// MockRoundTripper mocks HTTP downloads of oc binaries
type MockRoundTripper struct {
	delegate http.RoundTripper
}

func NewMockRoundTripper() http.RoundTripper {
	return &MockRoundTripper{DefaultRoundTripper}
}

func (t *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// for now only proxy the actual download requests
	re, _ := regexp.Compile(".*(openshift-origin-client-tools.*)&|.*(CHECKSUM).*")
	match := re.FindStringSubmatch(req.URL.String())

	if match != nil {
		filename := match[1]
		if filename == "" {
			filename = match[2]
		}
		//fmt.Printf("MockRoundTripper - Proxying request %s\n", req.URL.String() )

		response := &http.Response{
			Header:     make(http.Header),
			Request:    req,
			StatusCode: http.StatusOK,
		}
		response.Header.Set("Content-Type", "application/octet-stream")
		file, err := os.Open(filepath.Join(basepath, "..", "..", "test", "testdata", filename))
		if err != nil {
			panic(err)
		}

		response.Body = ioutil.NopCloser(bufio.NewReader(file))
		return response, nil
	}

	// otherwise delegate
	return t.delegate.RoundTrip(req)
}
