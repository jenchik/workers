package workers

type onDemand struct {
	g *Group
	w *Worker
}

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
