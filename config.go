package nexus

import "net/http"

type Config struct {
	ListenAddress       string
	Middlewares         []Middleware
	Endpoints           map[string][]Endpoint
	StaticFiles         http.FileSystem
	CrossOriginRequests CorsConfig
	Websockets          []Websocket
}
