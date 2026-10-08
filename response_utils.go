package nexus

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codeus-node/fail"
	di "github.com/codeus-node/generic-di"
)

func init() {
	di.Injectable(newResponseUtils)
}

type ResponseUtils interface {
	WritePlainText(status int, data string, w http.ResponseWriter) *HttpError
	WriteBytes(status int, data []byte, w http.ResponseWriter) *HttpError
	WriteJson(status int, data any, w http.ResponseWriter) *HttpError
	WriteEmpty(status int, w http.ResponseWriter) *HttpError
	WriteError(err *HttpError, w http.ResponseWriter)
}

type responseUtils struct{}

func (r *responseUtils) WriteEmpty(status int, w http.ResponseWriter) *HttpError {
	w.WriteHeader(status)
	_, err := w.Write([]byte{})
	if err != nil {
		return NewHttpError(fail.Wrap(err), http.StatusInternalServerError)
	}
	return nil
}

func (r *responseUtils) WriteError(err *HttpError, w http.ResponseWriter) {
	w.WriteHeader(err.Status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, innerErr := fmt.Fprintf(w, "Error in RequestTokenHandler: %s", err.Error.Error())
	if innerErr != nil {
		println(fmt.Sprintf("cannot write Error Response: %s", innerErr.Error()))
	}
}

func (r *responseUtils) WriteJson(status int, data any, w http.ResponseWriter) *HttpError {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return NewHttpError(fail.Wrap(err), http.StatusInternalServerError)
	}
	return nil
}

func (r *responseUtils) WritePlainText(status int, data string, w http.ResponseWriter) *HttpError {
	return printToStream(status, data, "text/plain; charset=utf-8", w)
}

func (r *responseUtils) WriteBytes(status int, data []byte, w http.ResponseWriter) *HttpError {
	return printToStream(status, data, "application/octet-stream", w)
}

func newResponseUtils() ResponseUtils {
	return &responseUtils{}
}

func printToStream[T any](status int, data T, contentType string, w http.ResponseWriter) *HttpError {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", contentType)
	_, err := fmt.Fprint(w, data)
	if err != nil {
		return NewHttpError(fail.Wrap(err), http.StatusInternalServerError)
	}
	return nil
}
