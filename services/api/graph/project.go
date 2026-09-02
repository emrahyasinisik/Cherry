package graph

import (
	"context"

	"github.com/icerde/api/internal/activate"
)

func (r *Resolver) projectPayload(ctx context.Context, userID, id string, full bool) (*Project, error) {
	row, err := r.Factory.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	out, err := mapProjectMeta(row)
	if err != nil {
		return nil, err
	}
	if !full {
		out.Logs = []*JobLog{}
		out.Files = []*ProjectFile{}
		out.Maestro = emptyMaestro()
		out.Activate = mapActivate(activate.Snapshot{Status: activate.StatusIdle, Note: "Yerel API kapalı."})
		return out, nil
	}
	logs, err := r.Factory.Logs(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	files, err := r.Factory.Files(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	studio, err := r.Factory.Maestro(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	maestro, err := mapMaestro(studio)
	if err != nil {
		return nil, err
	}
	out.Logs, err = mapLogs(logs)
	if err != nil {
		return nil, err
	}
	out.Files = mapFiles(files)
	out.Maestro = maestro
	out.Activate = mapActivate(r.Factory.ActivateSnap(id))
	return out, nil
}
