// Package restclient has a light implementation for doing 'curls' against HTTP endpoints.
// Not a big deal, but is makes the job easy.
package restclient

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru"
)

// Response is a struct that holds the information about the response of the call
type Response struct {
	Body          string
	Code          int
	Headers       map[string][]string
	CachedContent bool
	Staled        bool
}

// rest client (with cache) struct
type rClient struct {
	client  *http.Client
	baseURL string
	cache   *lru.Cache
	stale   bool
}

// PoolConfig is used to define a custom configuration for the pool
type PoolConfig struct {
	BaseURL             string
	MaxIdleConnsPerHost int
	Timeout             time.Duration
	Proxy               string
	CacheElements       int
	CacheState          bool
}

// Header is used to hold HTTP header values
type Header struct {
	Key   string
	Value string
}

type NotFollowRedirectError struct{}

func (e *NotFollowRedirectError) Error() string {
	return "Don't follow redirect"
}

// cache node
type cacheElement struct {
	Content string
	Headers map[string][]string
	Expires time.Time
}

// internal structure for mocks
type mockResponse struct {
	URL      string
	Method   string
	Response Response
	Headers  map[string]string
	Body     string
}

// connections pools
var pools = make(map[string]*rClient)

// list of mocks
var mocks []*mockResponse

// indicates if we have to use the mocks (good for testing)
var useMock bool

var notFollowRedirectError = new(NotFollowRedirectError)

// API constants
const (
	DefaultMaxIdleConnectionsPerHost = 100
)

// AddCustomPool creates a new connection pool based on the sent parameters
func AddCustomPool(pattern string, config *PoolConfig) {
	// create a transport for the connection
	transport := defaultTransport()

	// add the max per host config
	if config.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
	} else {
		transport.MaxIdleConnsPerHost = http.DefaultMaxIdleConnsPerHost
	}

	// sets the proxy
	if config.Proxy != "" {
		proxyURL, _ := url.Parse(config.Proxy)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// create the client
	client := &http.Client{Transport: transport}

	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		return notFollowRedirectError
	}

	if config.Timeout != 0 {
		client.Timeout = config.Timeout * time.Millisecond
	}

	rclient := new(rClient)
	rclient.client = client

	if config.BaseURL != "" {
		rclient.baseURL = config.BaseURL
	}

	if config.CacheElements > 0 {
		cache, _ := lru.New(config.CacheElements)
		rclient.cache = cache
		rclient.stale = config.CacheState
	}

	pools[pattern] = rclient
}

// Get executes a HTTP GET call to the specified url using headers to forward
func Get(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodGet, callURL, "", getHeadersMap(headers))
}

// Post executes a HTTP POST call to the specified url using headers to forward
func Post(callURL string, body string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodPost, callURL, body, getHeadersMap(headers))
}

// Put executes a HTTP PUT call to the specified url using headers to forward
func Put(callURL string, body string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodPut, callURL, body, getHeadersMap(headers))
}

// Delete executes a HTTP DELETE call to the specified url using headers to forward
func Delete(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodDelete, callURL, "", getHeadersMap(headers))
}

// Head executes a HTTP HEAD call to the specified url using headers to forward
func Head(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodHead, callURL, "", getHeadersMap(headers))
}

// Options executes a HTTP OPTIONS call to the specified url using headers to forward
func Options(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodOptions, callURL, "", getHeadersMap(headers))
}

// AddMocks adds seveal GET mocks for simple testing
func AddMocks(mocks map[string]Response) {
	// all all mocks in a bulk
	for key, value := range mocks {
		AddMock(key, http.MethodGet, "", value)
	}
}

// AddMock adds a URL, headers and Response to the mock URL map
func AddMock(URL string, method string, body string, response Response, headers ...Header) {
	// add the mock to the mock maps
	mResponse := new(mockResponse)
	mResponse.Response = response
	mResponse.Method = method
	mResponse.Headers = getHeadersMap(headers)
	mResponse.Body = body
	mResponse.URL = URL

	// create the map if doesnt exists of add the mock to the existing one
	if mocks == nil {
		mocks = make([]*mockResponse, 1)
		mocks[0] = mResponse
	} else {
		mocks = append(mocks, mResponse)
	}
}

// CleanMocks does what it says
func CleanMocks() {
	mocks = nil
}

// DisableMock disables all the URL mocks
func DisableMock() {
	useMock = false
}

// execute the request
func performRequest(method string, callURL string, body string, headers map[string]string) (*Response, error) {

	rclient := getPool(callURL)
	if rclient.baseURL != "" && !strings.Contains(callURL, rclient.baseURL) {
		callURL = rclient.baseURL + callURL
	}

	// if theere is a mock for the url and we are in testing, return the mock response
	if useMock {
		r := searchMockCall(method, callURL, headers, body)
		if r != nil {
			return r, nil
		}
	}

	// chech if we have to use the cache
	withCache := method == http.MethodGet && rclient.cache != nil

	var cachedResponse *Response

	if withCache {
		cachedResponse = getResponseFromCache(rclient, callURL)

		if cachedResponse != nil && !cachedResponse.Staled {
			return cachedResponse, nil
		}
	}

	var request *http.Request
	var error error

	// create the request to the API
	if method == http.MethodPost || method == http.MethodPut {
		request, error = http.NewRequest(method, callURL, bytes.NewBuffer([]byte(body)))

	} else {
		request, error = http.NewRequest(method, callURL, nil)
	}

	if error != nil {
		return nil, error
	}

	setHeaders(request, headers)
	response, error := rclient.client.Do(request)

	defer func() {
		// in case of error at opening, gives an error
		if response != nil {
			response.Body.Close()
		}
	}()

	var rcResponse *Response

	isNotFollowRedirectError := false

	if urlError, ok := error.(*url.Error); ok && urlError.Err == notFollowRedirectError {
		isNotFollowRedirectError = true

		error = nil
	}

	if error != nil && !isNotFollowRedirectError {
		rcResponse = &Response{"", 0, nil, false, false}
	}

	var byteBody []byte

	if rcResponse == nil {
		// read the response body
		if !isNotFollowRedirectError {
			byteBody, error = ioutil.ReadAll(response.Body)

			if error != nil {
				rcResponse = &Response{"", response.StatusCode, nil, false, false}
			}
		} else {
			byteBody = []byte("")
		}
	}

	if rcResponse == nil {
		rcResponse = &Response{string(byteBody), response.StatusCode, response.Header, false, false}
	}

	if withCache {
		// chek if we got 200
		if rcResponse.Code == http.StatusOK {
			setResponseInCache(rclient, rcResponse, callURL)

		} else {
			// if we got some error and the state option is configured, return the last good cached response
			if rclient.stale && cachedResponse != nil {
				return cachedResponse, nil
			}
		}
	}

	return rcResponse, error
}

func defaultTransport() *http.Transport {
	transport := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	return transport
}

// initDefaultPool initialize the rest client with the default settings
func initDefaultPool() *rClient {

	transport := defaultTransport()
	transport.MaxIdleConnsPerHost = DefaultMaxIdleConnectionsPerHost

	// create a http client to use, without timeout
	client := &http.Client{Transport: transport}

	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		return notFollowRedirectError
	}

	// creates the client-cache struct
	rclient := new(rClient)
	rclient.client = client

	pools["default"] = rclient

	return rclient
}

// return the http client based on the URL to call
func getPool(callURL string) *rClient {
	for pattern, pool := range pools {
		if regexp.MustCompile(pattern).MatchString(callURL) {
			return pool
		}
	}

	// create a default pool
	pool := pools["default"]
	if pool == nil {
		pool = initDefaultPool()
	}

	return pool
}

// setHeaders set the headers to the request
func setHeaders(request *http.Request, headers map[string]string) {
	// headers
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Connection", "Keep-Alive")

	// content-Type
	if request.Method != http.MethodGet {
		request.Header.Set("Content-Type", "application/json")
	}

	// forward the sent headers
	if headers != nil {
		for key, value := range headers {
			request.Header.Set(key, value)
		}
	}
}

// check if the headers map is equal
func sameHeaders(h1 map[string]string, h2 map[string]string) bool {
	// both nil
	if h1 == nil && h2 == nil {
		return true
	}

	// one nil
	if h1 == nil || h2 == nil {
		return false
	}

	// length
	if len(h1) != len(h2) {
		return false
	}

	// check every value and key in the other map
	for key, value := range h1 {

		// check if the key is in the other map
		val, ok := h2[key]
		if !ok {
			return false
		}

		// check the value
		if val != value {
			return false
		}

	}

	// if nothig was wrong return true (the maps are equal)
	return true
}

// returns the max age from the header
func getMaxAge(cacheControlValues []string) (int, error) {
	for _, value := range cacheControlValues {
		if strings.Contains(strings.ToLower(value), "max-age") {
			return strconv.Atoi(strings.Split(cacheControlValues[0], "=")[1])
		}
	}

	return 0, nil
}

// chechs if have to go to the cache for the element
func getResponseFromCache(rclient *rClient, callURL string) *Response {
	if rclient.cache.Contains(callURL) {
		// get the element from the cache
		element, _ := rclient.cache.Get(callURL)
		cacheElement := element.(*cacheElement)

		// if it is still valid, return the content from the cache
		if time.Now().Before(cacheElement.Expires) {
			return &Response{cacheElement.Content, 200, cacheElement.Headers, true, false}
		}

		// save the expired response for staled calls
		if rclient.stale {
			return &Response{cacheElement.Content, 200, cacheElement.Headers, true, true}
		}
	}

	return nil
}

// save the response to the cache
func setResponseInCache(rclient *rClient, response *Response, callURL string) {
	// checks the cache control header
	cacheControl := response.Headers["Cache-Control"]

	if cacheControl != nil {
		// check the max age value
		cacheControlValue, _ := getMaxAge(cacheControl)

		// checks if we have to cache or not
		if cacheControlValue > 0 {
			// create elemento to store the cache values
			cElement := new(cacheElement)
			cElement.Content = response.Body
			cElement.Headers = response.Headers
			cElement.Expires = time.Now().Add(time.Second * time.Duration(cacheControlValue))

			// save the data in the cache
			rclient.cache.Add(callURL, cElement)
		}
	}
}

func getHeadersMap(headers []Header) map[string]string {
	headersMap := make(map[string]string)

	for _, header := range headers {
		headersMap[header.Key] = header.Value
	}

	return headersMap
}

// search for the url in the mocks array
func searchMockCall(method string, callURL string, headers map[string]string, body string) *Response {
	for _, mock := range mocks {
		if mock.URL == callURL && mock.Method == method && sameHeaders(headers, mock.Headers) && body == mock.Body {
			return &mock.Response
		}
	}

	return nil
}

// init the Mock files
func init() {
	if os.Getenv("GO_ENVIRONMENT") != "production" {
		useMock = true
	}
}
