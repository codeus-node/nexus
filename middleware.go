package nexus

import "net/http"

const ExecuteNoMiddleware = "###NoMiddleware###"

type Middleware interface {
	GetKey() string
	Execute(w http.ResponseWriter, r *http.Request) *HttpError
}
