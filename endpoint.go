package nexus

import (
	"fmt"
	"net/http"

	"github.com/codeus-node/fail"
	di "github.com/codeus-node/generic-di"
	"github.com/codeus-node/printer"
)

type Endpoint interface {
	Methods() []string
	GetRoute() string
	ExplicitMiddlewares() []string
	Handler(http.ResponseWriter, *http.Request) *HttpError
	Before(http.ResponseWriter, *http.Request) *HttpError
	After(http.ResponseWriter, *http.Request) *HttpError
	OnError(err *HttpError, w http.ResponseWriter, r *http.Request)
}

type BaseEndpoint struct {
	logger        printer.Logger
	RequestUtils  RequestUtils
	ResponseUtils ResponseUtils
}

func NewBaseEndpoint() *BaseEndpoint {
	return &BaseEndpoint{
		logger:        di.Inject[printer.Logger](),
		RequestUtils:  di.Inject[RequestUtils](),
		ResponseUtils: di.Inject[ResponseUtils](),
	}
}

func (endpoint *BaseEndpoint) GetRoute() string {
	return "/"
}

func (endpoint *BaseEndpoint) Methods() []string {
	return []string{"GET"}
}

func (endpoint *BaseEndpoint) ExplicitMiddlewares() []string {
	return []string{}
}

func (endpoint *BaseEndpoint) Handler(http.ResponseWriter, *http.Request) *HttpError {
	return nil
}

func (endpoint *BaseEndpoint) Before(http.ResponseWriter, *http.Request) *HttpError {
	return nil
}

func (endpoint *BaseEndpoint) After(http.ResponseWriter, *http.Request) *HttpError {
	return nil
}

func (endpoint *BaseEndpoint) OnError(err *HttpError, w http.ResponseWriter, r *http.Request) {
	endpoint.logger.Log().PrintError(fail.Wrap(err.Error, fmt.Sprintf("with Status %d", err.Status)))
}
