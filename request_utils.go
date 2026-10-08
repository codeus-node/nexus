package nexus

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/codeus-node/fail"
	di "github.com/codeus-node/generic-di"
	"github.com/codeus-node/icons"
	"github.com/codeus-node/printer"
)

func init() {
	di.Injectable(newRequestUtils)
}

type RequestUtils interface {
	GetBearerToken(r *http.Request) string
	GetClientIP(r *http.Request) string
	GetPreferredLanguage(r *http.Request, supportedLanguages []string) string
	GetTraceID(r *http.Request) string
	GetHeaderValue(r *http.Request, name string) string
	GetUrlParam(r *http.Request, name string) string
	GetUrlParamInt(r *http.Request, name string, defaultValue int) int
	GetUrlParamBool(r *http.Request, name string, defaultValue bool) bool
	GetUrlParamFloat32(r *http.Request, name string, defaultValue float32) float32
	GetUrlParamFloat64(r *http.Request, name string, defaultValue float64) float64
	GetFormData(r *http.Request, name string) (string, fail.CustomError)
	GetRawBody(r *http.Request) ([]byte, fail.CustomError)
	GetJsonBody(r *http.Request, target any) fail.CustomError
}

type requestUtils struct {
	logger printer.Logger
}

func newRequestUtils() RequestUtils {
	return &requestUtils{
		logger: di.Inject[printer.Logger](),
	}
}

type langQ struct {
	lang string
	q    float64
}

func (ru *requestUtils) GetBearerToken(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

func (ru *requestUtils) GetUrlParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func (ru *requestUtils) GetUrlParamInt(r *http.Request, name string, defaultValue int) int {
	strValue := ru.GetUrlParam(r, name)
	v, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(v)
}

func (ru *requestUtils) GetUrlParamFloat64(r *http.Request, name string, defaultValue float64) float64 {
	strValue := ru.GetUrlParam(r, name)
	v, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return defaultValue
	}
	return v
}

func (ru *requestUtils) GetUrlParamFloat32(r *http.Request, name string, defaultValue float32) float32 {
	return float32(ru.GetUrlParamFloat64(r, name, float64(defaultValue)))
}

func (ru *requestUtils) GetUrlParamBool(r *http.Request, name string, defaultValue bool) bool {
	strValue := ru.GetUrlParam(r, name)
	if slices.Contains([]string{"true", "t", "yes", "ja"}, strings.ToLower(strValue)) {
		return true
	}
	if slices.Contains([]string{"false", "f", "no", "nein"}, strings.ToLower(strValue)) {
		return false
	}
	v, err := strconv.ParseBool(strValue)
	if err != nil {
		return defaultValue
	}
	return v
}

func (ru *requestUtils) GetFormData(r *http.Request, name string) (string, fail.CustomError) {
	if r.Method != http.MethodPost {
		return "", fail.Wrap(nil, "form data only support on POST Method")
	}

	return r.FormValue(name), nil
}

func (ru *requestUtils) GetHeaderValue(r *http.Request, name string) string {
	return r.Header.Get(name)
}

func (ru *requestUtils) GetRawBody(r *http.Request) ([]byte, fail.CustomError) {
	if r == nil || r.Body == nil {
		return nil, fail.Wrap(nil, "request or request body is nil")
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			ru.logger.Log().WithIcon(icons.Warning).PrintError(fail.Wrap(err, "can not close body from request"))
		}
	}()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fail.Wrap(err, "error on read request body")
	}

	return bodyBytes, nil
}

func (ru *requestUtils) GetJsonBody(r *http.Request, target any) fail.CustomError {
	stream, err := ru.GetRawBody(r)
	if err != nil {
		return err
	}

	unmarshalErr := json.Unmarshal(stream, target)
	if unmarshalErr != nil {
		return fail.Wrap(unmarshalErr, "error on read body stream into json format")
	}

	return nil
}

func (ru *requestUtils) GetClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return stripPort(ip)
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return stripPort(strings.TrimSpace(xri))
	}

	return stripPort(r.RemoteAddr)
}

func (ru *requestUtils) GetPreferredLanguage(r *http.Request, supportedLanguages []string) string {
	if r == nil || len(supportedLanguages) == 0 {
		return ""
	}

	header := r.Header.Get("Accept-Language")
	if header == "" {
		return supportedLanguages[0]
	}

	var preferred []langQ
	parts := strings.Split(header, ",")

	for _, part := range parts {
		subparts := strings.Split(strings.TrimSpace(part), ";")
		lang := strings.TrimSpace(subparts[0])
		if lang == "" {
			continue
		}

		qVal := 1.0
		if len(subparts) > 1 {
			qPart := strings.TrimSpace(subparts[1])
			if strings.HasPrefix(qPart, "q=") {
				if parsedQ, err := strconv.ParseFloat(qPart[2:], 64); err == nil {
					qVal = parsedQ
				}
			}
		}
		preferred = append(preferred, langQ{lang: lang, q: qVal})
	}

	sort.Slice(preferred, func(i, j int) bool {
		return preferred[i].q > preferred[j].q
	})

	for _, pref := range preferred {
		prefLang := strings.ToLower(pref.lang)

		for _, supported := range supportedLanguages {
			supportedLower := strings.ToLower(supported)

			if prefLang == supportedLower || strings.HasPrefix(prefLang, supportedLower+"-") {
				return supported
			}
		}
	}

	return supportedLanguages[0]
}

type contextKey string

const TraceIDKey contextKey = "trace_id"

func (ru *requestUtils) GetTraceID(r *http.Request) string {
	if r == nil {
		return ""
	}

	headersToSearch := []string{
		"X-Request-ID",
		"X-Trace-ID",
		"X-Correlation-ID",
		"Request-Id",
	}

	for _, h := range headersToSearch {
		if id := r.Header.Get(h); id != "" {
			return id
		}
	}

	if id, ok := r.Context().Value(TraceIDKey).(string); ok {
		return id
	}

	return ""
}

func stripPort(hostPort string) string {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return hostPort
	}
	return host
}
