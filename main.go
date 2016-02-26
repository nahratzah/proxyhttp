package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

type Proxy struct {
	client     *http.Client
	scheme     string
	targetHost *string // Optional
}

func NewProxyTargeted(scheme, targetHost string, client *http.Client) *Proxy {
	if client == nil {
		client = &http.Client{}
	}
	return &Proxy{
		scheme:     scheme,
		targetHost: &targetHost,
		client:     client,
	}
}

func NewProxy(scheme string, client *http.Client) *Proxy {
	if client == nil {
		client = &http.Client{}
	}
	return &Proxy{
		scheme:     scheme,
		targetHost: nil,
		client:     client,
	}
}

func (prx *Proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var url url.URL = *req.URL // struct copy
	url.Scheme = prx.scheme
	if prx.targetHost == nil {
		url.Host = req.Host
	} else {
		url.Host = *prx.targetHost
	}

	pReq, err := http.NewRequest(req.Method, url.String(), req.Body)
	if err != nil {
		resp.WriteHeader(503)
		return
	}

	// Copy headers
	for k, v := range req.Header {
		pReq.Header[k] = v
	}

	// No need to copy:
	// - Form, PostForm, MultipartForm: derived from body.
	// Need to copy:
	// - Trailer: trailer headers are announced in the headers and must be sent after the body.

	pResp, err := prx.client.Do(pReq)
	if err != nil {
		resp.WriteHeader(502)
		return
	}
	defer pResp.Body.Close()

	resp.WriteHeader(pResp.StatusCode)
	io.Copy(resp, pResp.Body)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", NewProxyTargeted("https", "www.google.com", nil))
	log.Fatal(http.ListenAndServe(":8000", mux))
}
