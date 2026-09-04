package graph

import "github.com/cherry/api/internal/store"

func (r *Resolver) connectionPayload(row store.Connection) (*Connection, error) {
	mode := "CONSENT"
	if r.Connect != nil {
		mode = r.Connect.ModeFor(row.Kind)
	}
	return mapConnectionWithMode(row, mode)
}
