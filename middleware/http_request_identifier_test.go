package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type RequestIdentifierTests struct {
	suite.Suite
}

func TestRequestIdentifier(t *testing.T) {
	suite.Run(t, &RequestIdentifierTests{})
}

func (s RequestIdentifierTests) Test_e2e_default() {
	expectedBodyText := "testing request identifier default"
	expectedHeaderKey := http.CanonicalHeaderKey(DefaultRequestIdentifierHeaderKey)
	c := RequestIdentifierConfiguration{}
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedBodyText)
	})
	withRequestIdentifier := NewRequestIdentifier(c)
	server := httptest.NewServer(withRequestIdentifier(handler))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	s.Nil(err)
	response, err := http.DefaultClient.Do(request)
	s.Nil(err)
	body, err := ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, expectedHeaderKey)
	_, err = uuid.Parse(response.Header.Get(expectedHeaderKey))
	s.Nil(err, "request identifier should be a valid uuid")
	s.Equal(expectedBodyText, string(body))
}

func (s RequestIdentifierTests) Test_e2e() {
	expectedBodyText := "testing request identifier"
	inputHeaderKey := "X-SOME-ID"
	expectedHeaderKey := "X-Some-Id"
	c := RequestIdentifierConfiguration{
		HeaderKey: inputHeaderKey,
	}
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedBodyText)
	})
	withRequestIdentifier := NewRequestIdentifier(c)
	server := httptest.NewServer(withRequestIdentifier(handler))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	s.Nil(err)
	response, err := http.DefaultClient.Do(request)
	s.Nil(err)
	body, err := ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, expectedHeaderKey)
	s.Equal(expectedBodyText, string(body))
}
