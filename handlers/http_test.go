package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/usvc/go-server/types"
)

type HandlersTests struct {
	suite.Suite
	latency time.Duration
}

func TestHandlers(t *testing.T) {
	suite.Run(t, &HandlersTests{
		latency: time.Millisecond,
	})
}

func (s HandlersTests) Test_GetHTTPLivenessProbe() {
	done := []bool{}
	livenessProbeHandlers := types.HTTPProbeHandlers{
		func() error {
			done = append(done, true)
			return nil
		},
		func() error {
			done = append(done, true)
			return nil
		},
	}
	livenessProbeHandler := GetHTTPLivenessProbe(livenessProbeHandlers)
	handler := http.NewServeMux()
	handler.HandleFunc("/", livenessProbeHandler)
	var testTasks sync.WaitGroup
	testTasks.Add(1)
	server := httptest.NewServer(handler)
	go func(after <-chan time.Time) {
		defer func() {
			server.Close()
			testTasks.Done()
		}()
		<-after
		request, err := http.NewRequest(http.MethodGet, server.URL, nil)
		s.Nil(err)
		response, err := http.DefaultClient.Do(request)
		s.Nil(err)
		body, err := ioutil.ReadAll(response.Body)
		s.Nil(err)
		s.Equal(http.StatusOK, response.StatusCode)
		s.Equal(ProbeResponseOK, string(body))
		s.Len(done, 2)
		s.True(done[0])
		s.True(done[1])
	}(time.After(s.latency))
	testTasks.Wait()

	expectedErrorMessage := "testing a liveness probe failure"
	livenessProbeHandlers = append(livenessProbeHandlers, func() error {
		return fmt.Errorf(expectedErrorMessage)
	})
	livenessProbeHandler = GetHTTPLivenessProbe(livenessProbeHandlers)
	handler = http.NewServeMux()
	handler.HandleFunc("/", livenessProbeHandler)
	testTasks.Add(1)
	server = httptest.NewServer(handler)
	go func(after <-chan time.Time) {
		defer func() {
			server.Close()
			testTasks.Done()
		}()
		<-after
		request, err := http.NewRequest(http.MethodGet, server.URL, nil)
		s.Nil(err)
		response, err := http.DefaultClient.Do(request)
		s.Nil(err)
		body, err := ioutil.ReadAll(response.Body)
		s.Nil(err)
		s.Contains(response.Header, "Content-Type")
		s.Equal("application/json", response.Header.Get("Content-Type"))
		s.Equal(http.StatusInternalServerError, response.StatusCode)
		s.Contains(string(body), expectedErrorMessage)
	}(time.After(s.latency))
	testTasks.Wait()
}

func (s HandlersTests) Test_GetHTTPMetricsProbe() {
	metricsHandler := GetHTTPMetrics()
	handler := http.NewServeMux()
	handler.HandleFunc("/", metricsHandler)
	server := httptest.NewServer(handler)
	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	s.Nil(err)
	response, err := http.DefaultClient.Do(request)
	s.Nil(err)
	body, err := ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(string(body), "promhttp_metric_handler_requests_total")

	expectedMetricName := "testing_get_http_metrics_probe"
	expectedMetricValue := float64(3.142)
	customRegistry := prometheus.NewRegistry()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: expectedMetricName,
		Help: "no need for this",
	})
	gauge.Set(expectedMetricValue)
	customRegistry.Register(gauge)
	metricsHandler = GetHTTPMetrics(customRegistry)
	handler = http.NewServeMux()
	handler.HandleFunc("/", metricsHandler)
	server = httptest.NewServer(handler)
	request, err = http.NewRequest(http.MethodGet, server.URL, nil)
	s.Nil(err)
	response, err = http.DefaultClient.Do(request)
	s.Nil(err)
	body, err = ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(string(body), fmt.Sprintf("%s %v", expectedMetricName, expectedMetricValue))
}

func (s HandlersTests) Test_GetHTTPReadinessProbe() {
	done := []bool{}
	readinessProbeHandlers := types.HTTPProbeHandlers{
		func() error {
			done = append(done, true)
			return nil
		},
		func() error {
			done = append(done, true)
			return nil
		},
	}
	readinessProbeHandler := GetHTTPReadinessProbe(readinessProbeHandlers)
	handler := http.NewServeMux()
	handler.HandleFunc("/", readinessProbeHandler)
	var testTasks sync.WaitGroup
	testTasks.Add(1)
	server := httptest.NewServer(handler)
	go func(after <-chan time.Time) {
		defer func() {
			server.Close()
			testTasks.Done()
		}()
		<-after
		request, err := http.NewRequest(http.MethodGet, server.URL, nil)
		s.Nil(err)
		response, err := http.DefaultClient.Do(request)
		s.Nil(err)
		body, err := ioutil.ReadAll(response.Body)
		s.Nil(err)
		s.Contains(response.Header, "Content-Type")
		s.Equal("application/json", response.Header.Get("Content-Type"))
		s.Equal(http.StatusOK, response.StatusCode)
		s.Equal(ProbeResponseOK, string(body))
		s.Len(done, 2)
		s.True(done[0])
		s.True(done[1])
	}(time.After(s.latency))
	testTasks.Wait()

	expectedErrorMessage := "testing a readiness probe failure"
	readinessProbeHandlers = append(readinessProbeHandlers, func() error {
		return fmt.Errorf(expectedErrorMessage)
	})
	readinessProbeHandler = GetHTTPReadinessProbe(readinessProbeHandlers)
	handler = http.NewServeMux()
	handler.HandleFunc("/", readinessProbeHandler)
	testTasks.Add(1)
	server = httptest.NewServer(handler)
	go func(after <-chan time.Time) {
		defer func() {
			server.Close()
			testTasks.Done()
		}()
		<-after
		request, err := http.NewRequest(http.MethodGet, server.URL, nil)
		s.Nil(err)
		response, err := http.DefaultClient.Do(request)
		s.Nil(err)
		body, err := ioutil.ReadAll(response.Body)
		s.Nil(err)
		s.Contains(response.Header, "Content-Type")
		s.Equal("application/json", response.Header.Get("Content-Type"))
		s.Equal(http.StatusInternalServerError, response.StatusCode)
		s.Contains(string(body), expectedErrorMessage)
	}(time.After(s.latency))
	testTasks.Wait()
}

func (s HandlersTests) Test_GetHTTPVersion() {
	expectedVersion := "testing http version"
	versionHandler := GetHTTPVersion(expectedVersion)
	handler := http.NewServeMux()
	handler.HandleFunc("/", versionHandler)
	server := httptest.NewServer(handler)
	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	s.Nil(err)
	response, err := http.DefaultClient.Do(request)
	s.Nil(err)
	body, err := ioutil.ReadAll(response.Body)
	s.Nil(err)
	s.Contains(response.Header, "Content-Type")
	s.Equal("text/plain", response.Header.Get("Content-Type"))
	s.Contains(string(body), expectedVersion)

}
