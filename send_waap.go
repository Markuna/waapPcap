package main

import (
	"bytes"
	"log"
	"net/http"
	// "sync/atomic"
)

// http发送客户端
var client = &http.Transport{}

// var counter uint64

func sendToWaap(oldReq *http.Request, body []byte) {
	// send to waap
	// atomic.AddUint64(&counter, 1)
	// log.Println(counter)
	url := *targetUrl + oldReq.RequestURI
	method := oldReq.Method
	payload := bytes.NewReader(body)

	request, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Println(err)
		return
	}
	request.Header = oldReq.Header

	response, err := client.RoundTrip(request)
	if err != nil {
		log.Println(err)
		return
	}
	// log.Println("send to waap end...")
	defer response.Body.Close()
}
