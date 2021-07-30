package httpclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	// HTTP methods we support
	POST    = "POST"
	GET     = "GET"
	HEAD    = "HEAD"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"

	HeaderContentType   = "Content-Type"
	ContentTypeJson     = "application/json"
	ContentTypeJsonUTF8 = "application/json;charset=UTF-8"
	ContentTypeForm     = "application/x-www-form-urlencoded"
	ContentTypeText     = "text/plain"
	ContentTypeXml      = "application/xml"

	DefaultTimeout               = 60 * time.Second
	DefaultDialTimeout           = 30 * time.Second
	DefaultKeepAliveTimeout      = 30 * time.Second
	DefaultIdleConnTimeout       = 90 * time.Second
	DefaultTLSHandshakeTimeout   = 10 * time.Second
	DefaultExpectContinueTimeout = 1 * time.Second
)

type CallBackStr func(response *http.Response, body string, err error)
type CallBack func(response *http.Response, err error)

type Client struct {
	url              string
	queries          map[string]string
	method           string
	header           map[string]string
	contentType      string
	body             []byte
	transport        *http.Transport
	timeout          time.Duration
	dialTimeout      time.Duration
	keepAliveTimeout time.Duration
	err              error
}

func (c *Client) DebugString() string {
	return fmt.Sprintf("[url]: %s\n"+
		"[method]: %s\n"+
		"[header]: %v\n"+
		"[content type]:%s\n"+
		"[body]:%s\n",
		c.getFullUrl(), c.method, c.header, c.contentType, c.body)
}

func (c *Client) getFullUrl() string {
	u, err := url.Parse(c.url)
	if err != nil {
		c.keepOriginErr(err)
		return err.Error()
	}
	query := u.Query()
	for k, v := range c.queries {
		query.Add(k, v)
	}
	u.RawQuery = query.Encode()
	return u.String()
}

func New() *Client {
	client := &Client{
		timeout:          DefaultTimeout,
		dialTimeout:      DefaultDialTimeout,
		keepAliveTimeout: DefaultKeepAliveTimeout,
		header:           make(map[string]string),
		transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			IdleConnTimeout:       DefaultIdleConnTimeout,
			TLSHandshakeTimeout:   DefaultTLSHandshakeTimeout,
			ExpectContinueTimeout: DefaultExpectContinueTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	return client
}

func (c *Client) Get(url string) *Client {
	return c.ReNew(GET, url)
}

func (c *Client) Post(url string) *Client {
	return c.ReNew(POST, url)
}

func (c *Client) Put(url string) *Client {
	return c.ReNew(PUT, url)
}

func (c *Client) Head(url string) *Client {
	return c.ReNew(HEAD, url)
}

func (c *Client) Delete(url string) *Client {
	return c.ReNew(DELETE, url)
}

func (c *Client) Patch(url string) *Client {
	return c.ReNew(PATCH, url)
}

func (c *Client) Options(url string) *Client {
	return c.ReNew(OPTIONS, url)
}

func (c *Client) AppendQuery(key, value string) *Client {
	if c.queries == nil {
		c.queries = map[string]string{key: value}
	} else {
		c.queries[key] = value
	}
	return c
}

func (c *Client) AppendQueries(queries map[string]string) *Client {
	if len(queries) == 0 {
		return c
	}
	if c.queries == nil {
		c.queries = make(map[string]string, len(queries))
	}
	for k, v := range queries {
		c.queries[k] = v
	}
	return c
}

func (c *Client) ReNew(method, url string) *Client {
	c.url = url
	c.method = method
	c.header = make(map[string]string)
	c.contentType = ""
	c.body = nil
	c.err = nil
	return c
}

func (c *Client) ContentType(contentType string) *Client {
	c.contentType = contentType
	return c
}

func (c *Client) Header(k, v string) *Client {
	if k == "" || v == "" {
		c.keepOriginErr(errors.New("invalid header, key or value is empty"))
	} else {
		c.header[k] = v
	}
	return c
}

// body can be defined struct, string, map, array(or slice), and so on
func (c *Client) Body(body interface{}) *Client {
	var err error
	switch value := body.(type) {
	case string:
		c.body = []byte(value)
	case []byte:
		c.body = value
	default:
		c.body, err = json.Marshal(body)
		c.keepOriginErr(err)
	}
	return c
}

func (c *Client) Timeout(timeout time.Duration) *Client {
	c.timeout = timeout
	return c
}

func (c *Client) DialTimeout(timeout time.Duration) *Client {
	c.dialTimeout = timeout
	return c
}

func (c *Client) KeepAliveTimeout(timeout time.Duration) *Client {
	c.keepAliveTimeout = timeout
	return c
}

func (c *Client) IdleConnTimeout(timeout time.Duration) *Client {
	c.transport.IdleConnTimeout = timeout
	return c
}

func (c *Client) TLSHandshakeTimeout(timeout time.Duration) *Client {
	c.transport.TLSHandshakeTimeout = timeout
	return c
}

func (c *Client) ExpectContinueTimeout(timeout time.Duration) *Client {
	c.transport.ExpectContinueTimeout = timeout
	return c
}

func (c *Client) keepOriginErr(err error) {
	if c.err == nil {
		c.err = err
	}
}

func (c *Client) Do(callback CallBack) {
	if callback == nil {
		return
	}
	callback(c.Go())
}

func (c *Client) Go() (*http.Response, error) {
	if c.err != nil {
		return nil, c.err
	}

	req, err := c.makeRequest()
	if err != nil {
		return nil, fmt.Errorf("make request failed:%q", err)
	}
	for k, v := range c.queries {
		req.URL.Query().Add(k, v)
	}
	client := c.makeClient()
	return client.Do(req)
}

func (c *Client) makeRequest() (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequest(c.method, c.getFullUrl(), bytes.NewReader(c.body))
	if err != nil {
		return nil, err
	}

	req.Header.Set(HeaderContentType, c.contentType)

	for k, v := range c.header {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (c *Client) makeClient() http.Client {
	c.transport.DialContext = (&net.Dialer{
		Timeout:   c.dialTimeout,
		KeepAlive: c.keepAliveTimeout,
	}).DialContext
	client := http.Client{
		Transport: c.transport,
		Timeout:   c.timeout,
	}
	return client
}
