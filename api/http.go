package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout:    60 * time.Second,
			DisableCompression: true,
		},
		Timeout: 60 * time.Second,
	}
}

func HttpRequest(client *http.Client, method, url string, headers map[string][]string, data, out, outErr interface{}) error {
	fmt.Printf("HTTP REQUEST %s %s\n", method, url)
	var body io.Reader
	if data != nil {
		spew.Dump(data)
		if b, ok := data.([]byte); ok {
			body = bytes.NewReader(b)
		} else {
			if b, err := json.Marshal(data); err == nil {
				body = bytes.NewReader(b)
			} else {
				return errors.New(fmt.Sprintf("Error marshaling request data: %v", err.Error()))
			}
		}
	}

	fmt.Println("Building request...")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building http request: %v", err.Error()))
	}
	for header, values := range headers {
		req.Header[header] = values
	}
	req.Header["content-type"] = []string{"application/json", "charset=utf-8"}
	spew.Dump(req.Header)

	fmt.Println("Sending request...")
	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error sending http request: %v", err.Error()))
	}
	fmt.Println("Req successful...")
	fmt.Println(resp.StatusCode)

	// normal response
	if resp.StatusCode == 200 && out != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading response body: %v", err.Error()))
		}
		if len(b) > 0 {
			log.Printf("Resp body: %s\n", string(b))
			return json.Unmarshal(b, out)
		}
	} else if resp.StatusCode != 200 {
		// error response
		var err error
		if outErr != nil {
			var b []byte
			b, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("Error reading err response body: %v", err.Error())
			}
			if len(b) > 0 {
				log.Printf("Err resp body: %s\n", string(b))
				if respErr := json.Unmarshal(b, outErr); respErr != nil {
					err = respErr
				}
			}
		}

		// should always throw an err if not 200
		if err == nil {
			err = errors.New("Network error")
		}
		return err
	}

	return nil
}

func HttpGetRequest(client *http.Client, method, url string, headers map[string][]string, data map[string]string, out interface{}) error {
	fmt.Printf("%s - %s\n", method, url)
	var body io.Reader

	// build request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building http request: %v", err.Error()))
	}
	for header, values := range headers {
		req.Header[header] = values
	}

	// add params
	if data != nil {
		q := req.URL.Query()
		for key, value := range data {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	spew.Dump(req.URL.RawQuery)

	// exec request
	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error sending http request: %v", err.Error()))
	}
	fmt.Printf("%s - %s\n", resp.Status, url)

	if out != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading response body: %v", err.Error()))
		}
		if len(b) > 0 {
			// fmt.Println(string(b))
			return json.Unmarshal(b, out)
		}
	}

	return nil
}

func HttpHtmlRequest(client *http.Client, method, url string, headers map[string][]string, data interface{}) (err error, out string) {
	var body io.Reader
	if data != nil {
		if b, ok := data.([]byte); ok {
			body = bytes.NewReader(b)
		} else {
			if b, err = json.Marshal(data); err == nil {
				body = bytes.NewReader(b)
			} else {
				err = errors.New(fmt.Sprintf("Error marshaling request data: %v", err.Error()))
				return
			}
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error building http request: %v", err.Error()))
		return
	}
	for header, values := range headers {
		req.Header[header] = values
	}

	resp, err := client.Do(req)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error sending http request: %v", err.Error()))
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error reading response body: %v", err.Error()))
		return
	}

	return nil, string(respBytes)
}
