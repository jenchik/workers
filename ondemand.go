package workers

type onDemand struct {
	g *Group
	w *Worker
}

// Run job in group context, wrap job to lock
func (d *onDemand) Run() error {
	if d.w.job == nil {
		return nil
	}
	select {
	case d.g.add <- d.w.RunOnce:
	case <-d.g.done:
		return ErrGroupStopped
	}
	return nil
}
