package main

import (
	"bytes"
	"crypto/tls"
	logit "github.com/cihub/seelog"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

/* Custom timout dialer, and the main reason is for avoiding closing the response before getting its body */
func timeoutDiall() *http.Client {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(netw, addr string) (net.Conn, error) {
			deadline := time.Now().Add(time.Duration(flusher_configuration.RequestTimeOut) * time.Millisecond)
			c, err := net.DialTimeout(netw, addr, time.Second)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(deadline)
			return c, nil
		}}
	httpclient := &http.Client{Transport: transport}
	return httpclient
}

/* Make request */
func callRequest(data string, url string, method string) ([]byte, *errorHandler) {
	client := timeoutDiall()
	req, req_err := http.NewRequest(method, url, bytes.NewBufferString(data))
	if req_err != nil {
		return nil, &errorHandler{req_err, req_err.Error(), 100020}
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	logit.Info(err)
	logit.Info(resp)
	if err != nil {
		return nil, &errorHandler{err, err.Error(), 100021}
	}

	defer resp.Body.Close()
	res_data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &errorHandler{err, err.Error(), 100022}
	}
	return res_data, nil
}
