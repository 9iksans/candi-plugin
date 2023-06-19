package echorest

import (
	"fmt"
	"net/http"
	"os"

	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/wrapper"
	"github.com/labstack/echo"
	"github.com/soheilhy/cmux"
)

const (
	// EchoREST types
	EchoREST types.Server = "echo-rest"
)

type (
	option struct {
		rootMiddlewares     []echo.MiddlewareFunc
		rootHandler         http.Handler
		errorHandler        echo.HTTPErrorHandler
		httpPort            uint16
		rootPath            string
		debugMode           bool
		includeGraphQL      bool
		jaegerMaxPacketSize int
		sharedListener      cmux.CMux
		engineOption        func(e *echo.Echo)
		graphqlOption       graphqlserver.Option
	}

	// OptionFunc type
	OptionFunc func(*option)
)

// CustomHTTPErrorHandler custom echo http error
func CustomHTTPErrorHandler(err error, c echo.Context) {
	var message string
	code := http.StatusInternalServerError
	if err != nil {
		message = err.Error()
	}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if code == http.StatusNotFound {
			message = fmt.Sprintf(`Resource "%s %s" not found`, c.Request().Method, c.Request().URL.Path)
		}
	}
	wrapper.NewHTTPResponse(code, message).JSON(c.Response())
}

func getDefaultOption() option {
	return option{
		httpPort:  8000,
		rootPath:  "",
		debugMode: true,
		rootMiddlewares: []echo.MiddlewareFunc{
			echo.WrapMiddleware(wrapper.HTTPMiddlewareCORS(
				env.BaseEnv().CORSAllowMethods, env.BaseEnv().CORSAllowHeaders,
				env.BaseEnv().CORSAllowOrigins, nil, env.BaseEnv().CORSAllowCredential,
			)),
			EchoWrapMiddleware(wrapper.HTTPMiddlewareTracer(wrapper.HTTPMiddlewareTracerConfig{
				MaxLogSize:  env.BaseEnv().JaegerMaxPacketSize,
				ExcludePath: map[string]struct{}{"/": {}, "/graphql": {}},
			})),
			EchoLoggerMiddleware(env.BaseEnv().DebugMode, os.Stdout),
		},
		rootHandler:  http.HandlerFunc(wrapper.HTTPHandlerDefaultRoot),
		errorHandler: CustomHTTPErrorHandler,
	}
}

// SetHTTPPort option func
func SetHTTPPort(port uint16) OptionFunc {
	return func(o *option) {
		o.httpPort = port
	}
}

// SetRootPath option func
func SetRootPath(rootPath string) OptionFunc {
	return func(o *option) {
		o.rootPath = rootPath
	}
}

// SetRootHTTPHandler option func
func SetRootHTTPHandler(rootHandler http.Handler) OptionFunc {
	return func(o *option) {
		o.rootHandler = rootHandler
	}
}

// SetSharedListener option func
func SetSharedListener(sharedListener cmux.CMux) OptionFunc {
	return func(o *option) {
		o.sharedListener = sharedListener
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *option) {
		o.debugMode = debugMode
	}
}

// SetIncludeGraphQL option func
func SetIncludeGraphQL(includeGraphQL bool) OptionFunc {
	return func(o *option) {
		o.includeGraphQL = includeGraphQL
	}
}

// SetJaegerMaxPacketSize option func
func SetJaegerMaxPacketSize(max int) OptionFunc {
	return func(o *option) {
		o.jaegerMaxPacketSize = max
	}
}

// SetRootMiddlewares option func
func SetRootMiddlewares(middlewares ...echo.MiddlewareFunc) OptionFunc {
	return func(o *option) {
		o.rootMiddlewares = middlewares
	}
}

// AddRootMiddlewares option func, overide root middleware
func AddRootMiddlewares(middlewares ...echo.MiddlewareFunc) OptionFunc {
	return func(o *option) {
		o.rootMiddlewares = append(o.rootMiddlewares, middlewares...)
	}
}

// SetErrorHandler option func
func SetErrorHandler(errorHandler echo.HTTPErrorHandler) OptionFunc {
	return func(o *option) {
		o.errorHandler = errorHandler
	}
}

// SetEchoEngineOption option func
func SetEchoEngineOption(echoFunc func(e *echo.Echo)) OptionFunc {
	return func(o *option) {
		o.engineOption = echoFunc
	}
}

// AddGraphQLOption option func
func AddGraphQLOption(opts ...graphqlserver.OptionFunc) OptionFunc {
	return func(o *option) {
		for _, opt := range opts {
			opt(&o.graphqlOption)
		}
	}
}
