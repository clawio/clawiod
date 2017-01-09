package registry

import (
	"context"
)

type Node struct {
	ID      string `json:"id"`
	Rol     string `json:"rol"`
	Version string `json:"version"`
	Host    string `json:"host"`
}

type Registry interface {
	Register(ctx context.Context, node *Node) error
	GetNodesForRol(ctx context.Context, rol string) ([]*Node, error)
}
