package restclient

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// start the testing webserver
	go restAPI()

	sleep()
	code := m.Run()

	os.Exit(code)
}

///// *** Test Cases *** /////

func TestGetDefault(t *testing.T) {

	// do the GET call
	response, err := Get("http://localhost:8080/testing")
	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("We didn't get 200: %v", response.Code)
	}

	// checks if the content of the body is as expected
	if response.Body != "{\"id\":\"123\"}" {
		t.Fatalf("The content was not as expected: %v", response.Body)
	}

	// cheks the headers
	if response.Headers == nil {
		t.Fatalf("Headers cant be nil")
	}

	// check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatalf("Content-Type: %s", contentType)
	}

	t.Log("End TestGetDefault")
}

func TestGetDefaultWithHeaders(t *testing.T) {
	//Do the GET call
	response, err := Get("http://localhost:8080/testing", Header{Key: "Test", Value: "Test Header"})

	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("We didn't get 200: %v", response.Code)
	}

	// checks if the content of the body is as expected
	if response.Body != "{\"id\":\"123\"}" {
		t.Fatalf("The content was not as expected: %v", response.Body)
	}

	// headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	// content of one header
	contentType := response.Headers["Test"]
	if contentType[0] != "Test Header" {
		t.Fatalf("The content was not as expected: %s", contentType[0])
	}

	t.Log("End TestGetDefaultWithHeaders")
}

func TestCustomPool(t *testing.T) {
	usersPool := new(PoolConfig)
	usersPool.BaseURL = "http://users.astropay.com"
	AddCustomPool("/users/.*", usersPool)

	cardsPool := new(PoolConfig)
	cardsPool.BaseURL = "http://cards.astropay.com"
	AddCustomPool("/cards/.*", cardsPool)

	// test pools
	client := getPool("http://users.astropay.com/users/1")
	if client.baseURL != usersPool.BaseURL {
		t.Fatalf("The content was not as expected, expected: %s, got: %s", usersPool.BaseURL, client.baseURL)
	}

	client = getPool("http://users.mercadolibre.com/cards/1234567890")
	if client.baseURL != cardsPool.BaseURL {
		t.Fatalf("The content was not as expected, expected: %s, got: %s", cardsPool.BaseURL, client.baseURL)
	}
}

func TestGetWithCustomPool(t *testing.T) {

	config := new(PoolConfig)
	config.MaxIdleConnsPerHost = 20
	config.Timeout = 0
	config.CacheElements = 0

	// create a new pool with 5ms of timeout
	AddCustomPool("http://localhost:8080", config)

	response, err := Get("http://localhost:8080/testing")
	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("We didn't get 200: %v", response.Code)
	}

	// checks if the content of the body is as expected
	if response.Body != "{\"id\":\"123\"}" {
		t.Fatalf("The content was not as expected: %v", response.Body)
	}

	// headers
	if response.Headers == nil {
		t.Fatalf("Headers cant be nil")
	}

	// content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatalf("Content-Type: %s", contentType)
	}

	t.Log("End TestGetWithCustomPool")
}

func TestGetWithPatternInCustomPool(t *testing.T) {

	config := new(PoolConfig)
	config.BaseURL = "http://localhost:8080"
	config.MaxIdleConnsPerHost = 20
	config.Timeout = 0
	config.CacheElements = 0

	// create a new pool with 5ms of timeout
	AddCustomPool("/testing.*", config)

	response, err := Get("/testing")
	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("We didn't get 200: %v", response.Code)
	}

	// checks body
	if response.Body != "{\"id\":\"123\"}" {
		t.Fatalf("The content was not as expected: %v", response.Body)
	}

	// headers
	if response.Headers == nil {
		t.Fatalf("Headers cant be nil")
	}

	//Check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatalf("Content-Type: %s", contentType)
	}

	//The end
	fmt.Println("End TestGetWithPatternInCustomPool")

}

func TestGetWithCustomPoolWithTimeout(t *testing.T) {

	config := new(PoolConfig)
	config.MaxIdleConnsPerHost = 20
	config.Timeout = 1

	AddCustomPool("http://localhost:8080", config)

	_, err := Get("http://localhost:8080/testing")
	if err == nil {
		t.Fatalf("We should had got a timeout: %s", err)
	}

	// restore the pool
	config.Timeout = 0
	AddCustomPool("http://localhost:8080", config)

	t.Log("End TestGetWithCustomPoolWithTimeout")
}

func TestGetWithCache(t *testing.T) {

	config := new(PoolConfig)
	config.CacheElements = 100

	AddCustomPool("http://localhost:8080", config)

	response, _ := Get("http://localhost:8080/cache?seconds=10")

	// checks the body
	if response.Body != "{\"id\":\"123\"}" || response.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	cachedResponse, _ := Get("http://localhost:8080/cache?seconds=10")

	// check the body
	if cachedResponse.Body != "{\"id\":\"123\"}" || cachedResponse.CachedContent == false {
		t.Log("Cached content:", cachedResponse.CachedContent)
		t.Log("Code:", cachedResponse.Code)
		t.Fatalf("The content was not as expected or was not cached: %s", cachedResponse.Body)
	}

	// test what happend if there is not Cache-Control header

	config.Timeout = 0
	config.CacheElements = 100
	AddCustomPool("http://localhost:8080", config)

	response, _ = Get("http://localhost:8080/cache")
	if response.Body != "{\"id\":\"123\"}" || response.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	// do the GET call again and check that the last call was not cached
	cachedResponse, _ = Get("http://localhost:8080/cache")
	if cachedResponse.Body != "{\"id\":\"123\"}" || cachedResponse.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", cachedResponse.Body)
	}

	// test what happend if the response has a Cache-Control=0
	config.Timeout = 0
	config.CacheElements = 100
	AddCustomPool("http://localhost:8080", config)

	// do the GET call (without cache control)
	response, _ = Get("http://localhost:8080/cache?seconds=0")
	if response.Body != "{\"id\":\"123\"}" || response.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	// do the GET call again and check that the last call was not cached
	cachedResponse, _ = Get("http://localhost:8080/cache?seconds=0")
	if cachedResponse.Body != "{\"id\":\"123\"}" || cachedResponse.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", cachedResponse.Body)
	}

	// reset the cache to (not cache)
	config.Timeout = 0
	AddCustomPool("http://localhost:8080", config)

	t.Log("End TestGetWithCache")
}

func TestGetWithStaleCache(t *testing.T) {

	config := new(PoolConfig)
	config.CacheElements = 100
	config.CacheState = true

	AddCustomPool("http://localhost:8080", config)

	response, _ := Get("http://localhost:8080/cache?seconds=1")
	if response.Body != "{\"id\":\"123\"}" || response.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	// do the GET call
	cachedResponse, _ := Get("http://localhost:8080/cache?seconds=1")
	if cachedResponse.Body != "{\"id\":\"123\"}" || cachedResponse.CachedContent == false || cachedResponse.Staled == true {
		fmt.Println("Cached content:", cachedResponse.CachedContent)
		fmt.Println("Code:", cachedResponse.Code)
		t.Fatalf("The content was not as expected or was not cached: %s", cachedResponse.Body)
	}

	// sleep for 2 seconds to expire the cachedResponse
	time.Sleep(1100 * time.Millisecond)

	// do the GET call and we hope to get the stale response (adding the header for 500 error response)
	staledResponse, _ := Get("http://localhost:8080/cache?seconds=1", Header{Key: "Error", Value: "500 error"})
	if staledResponse.Staled == false {
		fmt.Println("Cached content:", staledResponse.CachedContent)
		fmt.Println("Code:", staledResponse.Code)
		fmt.Println("Staled:", staledResponse.Staled)
		t.Fatalf("The content was not as expected or was not cached: %s", staledResponse.Body)
	}

	// reset the cache to (not cache)
	config.Timeout = 0
	AddCustomPool("http://localhost:8080", config)

	t.Log("End TestGetWithStaleCache")
}

func TestPostDefault(t *testing.T) {

	//Do the POST call
	response, err := Post("http://localhost:8080/testing", "{\"id\":\"123\"}")

	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 201 {
		t.Fatalf("There was not 201OK: %s", response.Code)
	}

	if response.Body != "echo --> {\"id\":\"123\"}" {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	// headers
	if response.Headers == nil {
		t.Fatalf("Headers cant be nil")
	}

	// check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatalf("Content-Type: %s", contentType)
	}

	t.Log("End TestPostDefault")
}

func TestPutDefault(t *testing.T) {

	//Do the PUT call
	response, err := Put("http://localhost:8080/testing", "{\"id\":\"123\"}")

	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("There was not 201OK: %s", response.Code)
	}

	if response.Body != "echoPut --> {\"id\":\"123\"}" {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	// headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	// content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatalf("Content-Type: %s", contentType)
	}

	t.Log("End TestPutDefault")
}

func TestDeleteDefault(t *testing.T) {

	//Do the Delete call
	response, err := Delete("http://localhost:8080/testing")

	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("There was not 200OK: %s", response.Code)
	}

	if response.Body != "echoDelete --> OK" {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	// headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	t.Log("End TestDeleteDefault")
}

func TestHeadDefault(t *testing.T) {

	//Do the Head call
	response, err := Head("http://localhost:8080/testing")

	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("There was not 200: %s", response.Code)
	}

	t.Log("End TestHeadDefault")
}

func TestMock(t *testing.T) {

	//Create the mock response
	mockResp := new(Response)
	mockResp.Body = "{\"id\":\"Pepe\"}"
	mockResp.CachedContent = false
	mockResp.Code = 200
	mockResp.Headers = nil
	mockResp.Staled = false

	AddMock("http://pepe.com", "GET", "", *mockResp)

	//Do the Get call
	response, err := Get("http://pepe.com")
	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("There was not 200: %s", response.Code)
	}

	if response.Body != "{\"id\":\"Pepe\"}" || response.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	t.Log("End TestMock")
}

func TestMockWithHeaders(t *testing.T) {

	headers := []Header{Header{Key: "Accept", Value: "application/json"}, Header{Key: "Encode", Value: "true"}}

	//Create the mock response
	mockResp := new(Response)
	mockResp.Body = "{\"id\":\"Pepe\"}"
	mockResp.CachedContent = false
	mockResp.Code = 200
	mockResp.Headers = nil
	mockResp.Staled = false

	AddMock("http://pepe.com", "GET", "", *mockResp, headers...)

	response, err := Get("http://pepe.com", headers...)
	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("There was not 200: %s", response.Code)
	}

	if response.Body != "{\"id\":\"Pepe\"}" || response.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	t.Log("End TestMockWithHeaders")
}

func TestSeveralGETMocks(t *testing.T) {

	mocks := make(map[string]Response)

	// create the mock response
	mocks["http://pepe.com"] = Response{Body: "{\"id\":\"Pepe\"}", Code: 200}

	AddMocks(mocks)

	// do the Get call
	response, err := Get("http://pepe.com")
	if err != nil {
		t.Fatalf("We got an error: %s", err)
	}

	if response.Code != 200 {
		t.Fatalf("We didn't get 200: %v", response.Code)
	}

	// checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"Pepe\"}" || response.CachedContent == true {
		t.Fatalf("The content was not as expected: %s", response.Body)
	}

	t.Log("End TestSeveralGETMocks")
}

///// *** Utils functions *** /////

// sleep for 100ms
func sleep() {
	time.Sleep(100 * time.Millisecond)
}

// init for testing
func restAPI() {
	// create the webserver
	http.Handle("/testing", http.HandlerFunc(processRequestDefault))
	http.Handle("/cache", http.HandlerFunc(processRequestCache))

	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
}

// processRequest process the request for the mock webserver
func processRequestCache(w http.ResponseWriter, req *http.Request) {

	// get the headers map
	headerMap := w.Header()

	parameter := req.URL.Query()["seconds"]
	if parameter != nil {
		seconds := req.URL.Query()["seconds"][0]

		if seconds != "" {
			headerMap.Add("Cache-Control", "max-age="+seconds)
		}
	}

	// copy headers sent to the response
	if req.Header["Error"] != nil {
		w.WriteHeader(500)
		w.Write([]byte("{\"error\":\"Error getting resource\"}"))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("{\"id\":\"123\"}"))
	}

	return

}

// processRequest process the request for the mock webserver
func processRequestDefault(w http.ResponseWriter, req *http.Request) {

	// get the headers map
	headerMap := w.Header()

	// returns the payload in the response (echo)
	headerMap.Add("Content-Type", "application/json;charset=UTF-8")

	// copy headers sent to the response
	if req.Header["Test"] != nil {
		headerMap.Add("Test", req.Header["Test"][0])
	}

	// performs action based on the request Method
	switch req.Method {

	case http.MethodGet:

		sleep()

		w.WriteHeader(200)
		w.Write([]byte("{\"id\":\"123\"}"))
		return

	case http.MethodHead:

		w.WriteHeader(200)
		return

	case http.MethodPut:

		// create the array to hold the body
		p := make([]byte, req.ContentLength)
		req.Body.Read(p)

		w.WriteHeader(200)
		w.Write([]byte("echoPut --> " + string(p)))

	case http.MethodDelete:
		w.WriteHeader(200)
		w.Write([]byte("echoDelete --> OK"))

	case http.MethodPost:

		// create the array to hold the body
		p := make([]byte, req.ContentLength)
		req.Body.Read(p)

		w.WriteHeader(201)
		w.Write([]byte("echo --> " + string(p)))

	default:
		// Method Not Allowed
		w.WriteHeader(405)
	}

}
