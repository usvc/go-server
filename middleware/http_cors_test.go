package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CORSTests struct {
	suite.Suite
}

func TestCORS(t *testing.T) {
	suite.Run(t, &CORSTests{})
}

func (s CORSTests) Test_e2e() {
	expectedBody := "testing hello hola"
	expectedAllowedHeaders := []string{"X-Test-Expected-One", "X-Test-Expected-Two"}
	expectedAllowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPatch}
	expectedAllowedOrigins := []string{"http://123.1.2.3", "http://123.2.3.4"}
	expectedExposedHeaders := []string{"X-Test-Exposed-One", "X-Test-Exposed-Two"}
	expectedMaxAge := time.Minute * 30
	expectedMaxAgeString := fmt.Sprintf("%v", expectedMaxAge.Seconds())
	c := CORSConfiguration{
		AllowCredentials:  true,
		AllowHeaders:      expectedAllowedHeaders,
		AllowMethods:      expectedAllowedMethods,
		AllowOrigins:      expectedAllowedOrigins,
		ExposeHeaders:     expectedExposedHeaders,
		EnablePassthrough: false,
		MaxAge:            expectedMaxAge,
	}
	withCORS := NewCORS(c)
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedBody)
	})
	server := httptest.NewServer(withCORS(handler))
	defer server.Close()

	// happy preflight request

	request, err := http.NewRequest(http.MethodOptions, server.URL, nil)
	request.Header.Add(CORSOrigin, expectedAllowedOrigins[0])
	request.Header.Add(CORSAccessControlRequestHeaders, expectedAllowedHeaders[0])
	request.Header.Add(CORSAccessControlRequestMethod, expectedAllowedMethods[0])
	s.Nil(err)
	response, err := http.DefaultClient.Do(request)
	s.Nil(err)
	body, err := ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, CORSAccessControlAllowCredentials)
	s.Equal("true", response.Header.Get(CORSAccessControlAllowCredentials))
	s.Contains(response.Header, CORSAccessControlAllowHeaders)
	s.Equal(expectedAllowedHeaders[0], response.Header.Get(CORSAccessControlAllowHeaders))
	s.Equal(expectedAllowedMethods[0], response.Header.Get(CORSAccessControlAllowMethods))
	s.Contains(response.Header, CORSAccessControlAllowOrigin)
	s.Equal(expectedAllowedOrigins[0], response.Header.Get(CORSAccessControlAllowOrigin))
	s.Equal(expectedMaxAgeString, response.Header.Get(CORSAccessControlMaxAge))
	s.Equal(http.StatusNoContent, response.StatusCode)
	s.Equal("", string(body))

	// sad preflight request (origin)

	request, err = http.NewRequest(http.MethodOptions, server.URL, nil)
	request.Header.Add(CORSOrigin, "http://unexpectedorigin.com")
	request.Header.Add(CORSAccessControlRequestHeaders, expectedAllowedHeaders[0])
	request.Header.Add(CORSAccessControlRequestMethod, expectedAllowedMethods[0])
	s.Nil(err)
	response, err = http.DefaultClient.Do(request)
	s.Nil(err)
	body, err = ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, CORSAccessControlAllowCredentials)
	s.Equal("true", response.Header.Get(CORSAccessControlAllowCredentials))
	s.Contains(response.Header, CORSAccessControlAllowHeaders)
	s.Equal(expectedAllowedHeaders[0], response.Header.Get(CORSAccessControlAllowHeaders))
	s.Equal(expectedAllowedMethods[0], response.Header.Get(CORSAccessControlAllowMethods))
	s.NotContains(response.Header, CORSAccessControlAllowOrigin)
	s.Equal(http.StatusBadRequest, response.StatusCode)
	s.Equal("", string(body))

	// sad preflight request (headers)

	request, err = http.NewRequest(http.MethodOptions, server.URL, nil)
	request.Header.Add(CORSOrigin, expectedAllowedOrigins[0])
	request.Header.Add(CORSAccessControlRequestHeaders, "X-Unexpected-Header")
	request.Header.Add(CORSAccessControlRequestMethod, expectedAllowedMethods[0])
	s.Nil(err)
	response, err = http.DefaultClient.Do(request)
	s.Nil(err)
	body, err = ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, CORSAccessControlAllowCredentials)
	s.Equal("true", response.Header.Get(CORSAccessControlAllowCredentials))
	s.NotContains(response.Header, CORSAccessControlAllowHeaders)
	s.Contains(response.Header, CORSAccessControlAllowMethods)
	s.Equal(expectedAllowedMethods[0], response.Header.Get(CORSAccessControlAllowMethods))
	s.Contains(response.Header, CORSAccessControlAllowOrigin)
	s.Equal(expectedAllowedOrigins[0], response.Header.Get(CORSAccessControlAllowOrigin))
	s.Equal(http.StatusBadRequest, response.StatusCode)
	s.Equal("", string(body))

	// sad preflight request (methods)

	request, err = http.NewRequest(http.MethodOptions, server.URL, nil)
	request.Header.Add(CORSOrigin, expectedAllowedOrigins[0])
	request.Header.Add(CORSAccessControlRequestHeaders, expectedAllowedHeaders[0])
	request.Header.Add(CORSAccessControlRequestMethod, http.MethodDelete)
	s.Nil(err)
	response, err = http.DefaultClient.Do(request)
	s.Nil(err)
	body, err = ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, CORSAccessControlAllowCredentials)
	s.Equal("true", response.Header.Get(CORSAccessControlAllowCredentials))
	s.Contains(response.Header, CORSAccessControlAllowHeaders)
	s.Equal(expectedAllowedHeaders[0], response.Header.Get(CORSAccessControlAllowHeaders))
	s.NotContains(response.Header, CORSAccessControlAllowMethods)
	s.Contains(response.Header, CORSAccessControlAllowOrigin)
	s.Equal(expectedAllowedOrigins[0], response.Header.Get(CORSAccessControlAllowOrigin))
	s.Equal(http.StatusBadRequest, response.StatusCode)
	s.Equal("", string(body))

	// actual request

	request, err = http.NewRequest(http.MethodGet, server.URL, nil)
	request.Header.Add(CORSAccessControlRequestHeaders, expectedAllowedHeaders[0])
	request.Header.Add(CORSAccessControlRequestHeaders, expectedAllowedHeaders[1])
	request.Header.Add(CORSOrigin, expectedAllowedOrigins[0])
	s.Nil(err)
	response, err = http.DefaultClient.Do(request)
	s.Nil(err)
	body, err = ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, CORSAccessControlAllowCredentials)
	s.Equal("true", response.Header.Get(CORSAccessControlAllowCredentials))
	s.Equal(expectedAllowedHeaders[0], response.Header.Get(CORSAccessControlAllowHeaders))
	s.Equal(http.MethodGet, response.Header.Get(CORSAccessControlAllowMethods))
	s.Equal(expectedAllowedOrigins[0], response.Header.Get(CORSAccessControlAllowOrigin))
	s.Equal(strings.Join(expectedExposedHeaders, ","), response.Header.Get(CORSAccessControlExposeHeaders))
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal(expectedBody, string(body))
}
