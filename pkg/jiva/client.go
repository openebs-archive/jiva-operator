// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jiva

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type ControllerClient struct {
	Address    string
	httpClient *http.Client
}

// NewControllerClient create the new controller client
func NewControllerClient(address string) *ControllerClient {
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	if !strings.HasSuffix(address, "/v1") {
		address += "/v1"
	}

	return &ControllerClient{
		Address:    address,
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}
}

// Get sends a request to the specified path and stores body in the value
// pointed to by obj.
func (c *ControllerClient) Get(path string, obj interface{}) error {
	resp, err := http.Get(c.Address + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}

// Post sends a POST request to the specified path and stores body in the value
// pointed to by resp.
func (c *ControllerClient) Post(path string, req, resp interface{}) error {
	return c.Do("POST", path, req, resp)
}

// Do sends a request to the specified path and it stores JSON-decoded body
// from the response into the value pointed to by resp.
func (c *ControllerClient) Do(method, path string, req, resp interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	bodyType := "application/json"
	url := path
	if !strings.HasPrefix(url, "http") {
		url = c.Address + path

	}

	httpReq, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", bodyType)

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 300 {
		content, _ := ioutil.ReadAll(httpResp.Body)
		return fmt.Errorf("Bad response: %d %s: %s", httpResp.StatusCode, httpResp.Status, content)
	}

	if resp == nil {
		return nil
	}
	return json.NewDecoder(httpResp.Body).Decode(resp)
}
