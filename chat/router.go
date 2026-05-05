package chat

import "fmt"

// Router resolves a named role to a concrete Client.
// Roles are task-defined strings like "translate", "proofread", "summarize".
type Router interface {
	Resolve(role string) (Client, error)
}

// SingleClient returns a Router that always resolves to the same client
// regardless of the role requested.
func SingleClient(c Client) Router {
	return &singleRouter{client: c}
}

type singleRouter struct {
	client Client
}

func (r *singleRouter) Resolve(_ string) (Client, error) {
	return r.client, nil
}

// Map returns a Router backed by explicit role-to-client mappings.
// Unknown roles fall back to defaultClient if non-nil.
func Map(clients map[string]Client, defaultClient Client) Router {
	return &mapRouter{clients: clients, defaultClient: defaultClient}
}

type mapRouter struct {
	clients       map[string]Client
	defaultClient Client
}

func (r *mapRouter) Resolve(role string) (Client, error) {
	if c, ok := r.clients[role]; ok {
		return c, nil
	}
	if r.defaultClient != nil {
		return r.defaultClient, nil
	}
	return nil, fmt.Errorf("no client registered for role %q", role)
}
