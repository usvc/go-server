package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RequestLoggerTest struct {
	suite.Suite
}

func TestRequestLogger(t *testing.T) {
	suite.Run(t, &RequestLoggerTest{})
}

func (s RequestLoggerTest) Test_e2e() {
	expectedBodyText := "testing request logger"
	expectedURI := "/expected/uri?with=query&params=true"
	expectedUserAgent := "expected-user-agent/1.2.3"
	expectedHeaderKey := "X-Header"
	expectedHeaderValue := "some random header"
	var output bytes.Buffer
	c := RequestLoggerConfiguration{Log: func(args ...interface{}) {
		fmt.Fprint(&output, args...)
	}}
	withRequestLogger := NewRequestLogger(c)
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(expectedHeaderKey, expectedHeaderValue)
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(expectedBodyText))
	})
	server := httptest.NewServer(withRequestLogger(handler))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+expectedURI, bytes.NewBufferString("hello world"))
	request.Header.Set("User-Agent", expectedUserAgent)
	request.SetBasicAuth("username", "password")
	s.Nil(err)
	response, err := http.DefaultClient.Do(request)
	s.Nil(err)
	body, err := ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Equal(http.StatusTeapot, response.StatusCode)
	s.Equal(expectedBodyText, string(body))
	s.Contains(response.Header, expectedHeaderKey)
	s.Equal(expectedHeaderValue, response.Header.Get(expectedHeaderKey))
	logEntry := string(output.Bytes())
	s.Contains(logEntry, fmt.Sprintf("GET %s", expectedURI),
		"the request uri should be logged")
	s.Contains(logEntry, expectedUserAgent,
		"the user agent should be logged")
	s.Regexp(`^[0-9\.\:]+[0-9]+`, logEntry,
		"the remote address should be logged")
	s.Regexp(`[0-9]+/[A-Z][a-z]{2}/[0-9]{4}:[0-9]{2}:[0-9]{2}:[0-9]{2} \+0000`, logEntry,
		"the request timestamp should be logged in utc")
	s.Regexp(`rt=0\.[0-9]+`, logEntry,
		"the request latency should be logged")
	s.Regexp(`id=-`, logEntry,
		"the request id should be logged if its available")
}

func (s RequestLoggerTest) Test_formatInterface() {
	testInterface := interface{}(-1)
	s.Equal("-1", formatInterface(testInterface), "shoud parse integers")
	testInterface = interface{}(uint(1))
	s.Equal("1", formatInterface(testInterface), "shoud parse unsigned integers")
	testInterface = interface{}(float64(3.142))
	s.Equal("3.142", formatInterface(testInterface), "shoud parse floating points")
	testInterface = interface{}("a")
	s.Equal("a", formatInterface(testInterface), "shoud parse strings")
	testInterface = interface{}(true)
	s.Equal("true", formatInterface(testInterface), "shoud parse booleans")
	testInterface = interface{}(nil)
	s.Equal("-", formatInterface(testInterface), "shoud parse nil")

}
