# `go-server`

[![go version](https://img.shields.io/github/go-mod/go-version/usvc/go-server)](https://github.com/usvc/go-server)
[![tag github](https://img.shields.io/github/v/tag/usvc/go-server)](https://github.com/usvc/go-server/tags)
[![build status](https://travis-ci.org/usvc/go-server.svg?branch=master)](https://travis-ci.org/usvc/go-server)
[![pipeline status](https://gitlab.com/usvc/modules/go/server/badges/master/pipeline.svg)](https://gitlab.com/usvc/modules/go/server/-/commits/master)
[![Test Coverage](https://api.codeclimate.com/v1/badges/acfc321def02f47b57a2/test_coverage)](https://codeclimate.com/github/usvc/go-server/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/acfc321def02f47b57a2/maintainability)](https://codeclimate.com/github/usvc/go-server/maintainability)

A Go package to deal with setting up a cloud-native microservices-ready server

|        |                                                                                        |
| ------ | -------------------------------------------------------------------------------------- |
| Github | [https://github.com/usvc/go-server](https://github.com/usvc/go-server)                 |
| Gitlab | [https://gitlab.com/usvc/modules/go/server](https://gitlab.com/usvc/modules/go/server) | . |

## Why use this

This server package comes with the following in-built and configured with reasonable defaults:

1. Prometheus metrics
2. Healthcheck probes
3. Request identification
4. Request logging
5. Cross-Origin Resource Sharing (CORS)
6. Graceful shutdown handling

## Usage

### Importing

```go
import "github.com/usvc/go-server"
```

### Creating a server

```go
package main

import (
  "net/http"
  "github.com/usvc/go-server"
)

func main() {
  options := server.NewHTTPOptions()
  mux := http.NewServeMux()
  // ... other configuration tasks ...
	s := server.NewHTTP(options, mux)
}
```

### Using a custom logger

```go
// ...
  options := server.NewHTTPOptions()
  options.Loggers.ServerEvent = func(args ...interface{}) {
    logrus.Debug(args...)
  }
  options.Loggers.Request = func(args ...interface{}) {
    logrus.Trace(args...)
  }
// ...
```

### Using liveness/readiness probes

Both types of healthchecks implement a similar pattern

```go
// ...
  options := server.NewHTTPOptions()
  // for liveness probes
  options.LivenessProbe.Handlers = types.HTTPProbeHandlers{
    func() error {
      // ... some checks maybe? ...
      return nil
    },
  }
  // for readiness probes
  options.ReadinessProbe.Handlers = types.HTTPProbeHandlers{
    func() error {
      // ... some checks maybe? ...
      return nil
    },
  }
// ...
```

### Using a custom path for probes/metrics

```go
// ...
  options := server.NewHTTPOptions()
  // ... use /not-healthz as the liveness probe endpoint ...
  options.LivenessProbe.Path = "/not-healthz"
  // ... use /see-whats-inside as the metrics endpoint ...
  options.Metrics.Path = "/see-whats-inside"
  // ... use /not-healthz as the readiness probe endpoint ...
  options.ReadinessProbe.Path = "/not-readyz"
// ...
```

### Password protecting provided paths

```go
// ...
  options := server.NewHTTPOptions()
  // ... protect the liveness probe endpoint with a password...
  options.LivenessProbe.Password = "123456"
  // ... protect the metrics endpoint with a password...
  options.Metrics.Password = "123456"
  // ... protect the readiness probe endpoint with a password...
  options.ReadinessProbe.Password = "123456"
// ...
```

### Using custom middlewares

```go
// ...
  options := server.NewHTTPOptions()
  options.Middlewares = append(options.Middlewares, func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      // ... custom middleware logic ...
      next.ServeHTTP(w, r)
    })
  })
// ...
```

### Disabling features

```go
// ...
  options := server.NewHTTPOptions()
  
  // to disable CORS
  options.Disable.CORS = false

  // to disable the liveness probe endpoint from being registered
  options.Disable.LivenessProbe = false

  // to disable the metrics endpoint from being reigstered
  options.Disable.Metrics = false

  // to disable the readiness probe endpoint from being registered
  options.Disable.ReadinessProbe = false

  // to disable the request identification middleware
  options.Disable.RequestIdentifier = false

  // to disable the request logging middleware
  options.Disable.RequestLogger = false

  // to disable the syscall signal handler middleware
  options.Disable.SignalHandling = false

  // to disable the version endpoint from being registered
  options.Disable.Version = false
// ...
```

## Development Runbook

### Getting Started

1. Clone this repository
2. Run `make deps` to pull in external dependencies
3. Write some awesome stuff
4. Run `make test` to ensure unit tests are passing
5. Push

### Continuous Integration (CI) Pipeline

#### On Github

Github is used to deploy binaries/libraries because of it's ease of access by other developers.

##### Releasing

Releasing of the binaries can be done via Travis CI.

1. On Github, navigate to the [tokens settings page](https://github.com/settings/tokens) (by clicking on your profile picture, selecting **Settings**, selecting **Developer settings** on the left navigation menu, then **Personal Access Tokens** again on the left navigation menu)
2. Click on **Generate new token**, give the token an appropriate name and check the checkbox on **`public_repo`** within the **repo** header
3. Copy the generated token
4. Navigate to [travis-ci.org](https://travis-ci.org) and access the cooresponding repository there. Click on the **More options** button on the top right of the repository page and select **Settings**
5. Scroll down to the section on **Environment Variables** and enter in a new **NAME** with `RELEASE_TOKEN` and the **VALUE** field cooresponding to the generated personal access token, and hit **Add**

#### On Gitlab

Gitlab is used to run tests and ensure that builds run correctly.

##### Version Bumping

1. Run `make .ssh`
2. Copy the contents of the file generated at `./.ssh/id_rsa.base64` into an environment variable named **`DEPLOY_KEY`** in **Settings > CI/CD > Variables**
3. Navigate to the **Deploy Keys** section of the **Settings > Repository > Deploy Keys** and paste in the contents of the file generated at `./.ssh/id_rsa.pub` with the **Write access allowed** checkbox enabled

- **`DEPLOY_KEY`**: generate this by running `make .ssh` and copying the contents of the file generated at `./.ssh/id_rsa.base64`

##### DockerHub Publishing

1. Login to [https://hub.docker.com](https://hub.docker.com), or if you're using your own private one, log into yours
2. Navigate to [your security settings at the `/settings/security` endpoint](https://hub.docker.com/settings/security)
3. Click on **Create Access Token**, type in a name for the new token, and click on **Create**
4. Copy the generated token that will be displayed on the screen
5. Enter the following varialbes into the CI/CD Variables page at **Settings > CI/CD > Variables** in your Gitlab repository:

- **`DOCKER_REGISTRY_URL`**: The hostname of the Docker registry (defaults to `docker.io` if not specified)
- **`DOCKER_REGISTRY_USERNAME`**: The username you used to login to the Docker registry
- **`DOCKER_REGISTRY_PASSWORD`**: The generated access token

## Licensing

Code in this package is licensed under the [MIT license (click to see full text))](./LICENSE)
