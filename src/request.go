package direwolf

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Request is a prepared request that generated by prepareRequest().
// It came from your input in Request.
type Request struct {
	Method   string
	URL      string
	Headers  http.Header
	Data     Data
	DataForm url.Values
	Params   url.Values
	Cookies  []*http.Cookie
}

// setHeader get the key-value from Headers to Request.Headers.
func (req *Request) setHeader(h Headers) {
	req.Headers = http.Header{}
	for key, slice := range h {
		for _, value := range slice {
			req.Headers.Add(key, value)
		}
	}
}

// setParams set Request.Params.Encode Params and join it to url.
func (req *Request) setParams(p Params) {
	req.Params = url.Values(p)
	req.URL = req.URL + "?" + req.Params.Encode() // add params to url
}

// setCookies set Request.Cookies
func (req *Request) setCookies(c Cookies) {
	req.Cookies = []*http.Cookie{}
	for key, value := range c {
		req.Cookies = append(req.Cookies, &http.Cookie{Name: key, Value: value})
	}
}

// Session is the main object in direwolf. This is its main features:
// 1. handling redirects
// 2. automatically managing cookies
type Session struct {
	client *http.Client
}

// prepareRequest is to process the parameters from user input.Generate PreRequest object.
func (session Session) prepareRequest(method string, URL string, args ...interface{}) *Request {
	req := new(Request)
	req.Method = method
	req.URL = URL

	// Check the type of the paramter and handle it.
	for _, arg := range args {
		switch a := arg.(type) {
		case Headers:
			req.setHeader(a)
		case http.Header:
			req.Headers = a
		case Params:
			req.setParams(a)
		case DataForm:
			req.DataForm = url.Values(a)
		case Data:
			req.Data = a
		case Cookies:
			req.setCookies(a)
		}
	}
	return req
}

// Request is a generic request method.
func (session *Session) request(method string, URL string, args ...interface{}) {
	preq := session.prepareRequest(method, URL, args...)
	session.send(preq)
}

// Get is a get method.
func (session *Session) Get(URL string, args ...interface{}) {
	session.request("Get", URL, args...)
}

// Post is a post method.
func (session *Session) Post(URL string, args ...interface{}) {
	session.request("Post", URL, args...)
}

// send is responsible for handling some subsequent processing of the PreRequest.
func (session *Session) send(preq *Request) *Response {
	session.client = &http.Client{}
	req, err := http.NewRequest(preq.Method, preq.URL, nil)
	if err != nil {
		panic(err)
	}

	// Handle the Headers.
	req.Header = preq.Headers
	// Handle the DataForm, convert DataForm to strings.Reader.
	// add two new headers: Content-Type and ContentLength.
	if preq.DataForm != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		data := preq.DataForm.Encode()
		req.Body = ioutil.NopCloser(strings.NewReader(data))
		req.ContentLength = int64(len(data))
	}
	// Handle Cookies
	if preq.Cookies != nil {
		for _, cookie := range preq.Cookies {
			req.AddCookie(cookie)
		}
	}

	resp, err := session.client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	return &Response{}
}