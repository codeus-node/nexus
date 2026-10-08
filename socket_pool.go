package nexus

import (
	"sync"

	di "github.com/codeus-node/generic-di"
)

func init() {
	di.Injectable(newSocketPool)
}

type ClientPool interface {
	Register(client WebsocketClient)
	Get(filter func(client WebsocketClient) bool) []WebsocketClient
}

type clientPool struct {
	mutex sync.Mutex
	store map[string]WebsocketClient
}

func newSocketPool() ClientPool {
	return &clientPool{
		mutex: sync.Mutex{},
		store: make(map[string]WebsocketClient),
	}
}

func (pool *clientPool) Register(client WebsocketClient) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.store[client.GetId()] = client
}

func (pool *clientPool) Get(filter func(client WebsocketClient) bool) []WebsocketClient {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	result := make([]WebsocketClient, 0)
	for _, socket := range pool.store {
		if filter(socket) {
			result = append(result, socket)
		}
	}
	return result
}
