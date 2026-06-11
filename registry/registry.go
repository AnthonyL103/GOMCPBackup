package registry

import (
	"fmt"
	"github.com/AnthonyL103/GOMCP/server"
)

type Registry struct {
	Servers map[string]*server.MCPServer
}

func NewRegistry() *Registry {
	return &Registry{
		Servers: make(map[string]*server.MCPServer),
	}
}

// AddServer adds a server to the registry
func (r *Registry) AddServer(srv *server.MCPServer) error {
	if srv == nil {
		return fmt.Errorf("server cannot be nil")
	}

	if _, exists := r.Servers[srv.ServerID]; exists {
		return fmt.Errorf("server with ID '%s' already exists", srv.ServerID)
	}

	r.Servers[srv.ServerID] = srv
	return nil
}

// RemoveServer removes a server from the registry
func (r *Registry) RemoveServer(serverID string) error {
	if _, exists := r.Servers[serverID]; !exists {
		return fmt.Errorf("server with ID '%s' does not exist", serverID)
	}

	delete(r.Servers, serverID)
	return nil
}

// GetServer retrieves a server from the registry
func (r *Registry) GetServer(serverID string) (*server.MCPServer, error) {
	if srv, exists := r.Servers[serverID]; exists {
		return srv, nil
	}

	return nil, fmt.Errorf("server with ID '%s' does not exist", serverID)
}

// ListServers returns all servers in the registry
func (r *Registry) ListServers() []*server.MCPServer {
	servers := make([]*server.MCPServer, 0, len(r.Servers))
	for _, srv := range r.Servers {
		servers = append(servers, srv)
	}
	return servers
}

// ExecuteTool executes a tool from a specific server
/*
func (r *Registry) ExecuteTool(serverID, toolID string, input map[string]interface{}) (interface{}, error) {
	srv, err := r.GetServer(serverID)
	if err != nil {
		return nil, err
	}

	return srv.ExecuteTool(toolID, input)
}
*/