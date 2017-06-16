package pipkin

import "context"

type Process struct {
	ctx   context.Context
	units []*Unit

	cancel func()
}

func (p *Process) Use(unit ...*Unit) {
	max := uint8(0)
	for _, u := range unit {
		if u.id > max {
			max = u.id
		}
	}
	if len(p.units) < int(max+1) {
		us := make([]*Unit, max+1)
		copy(us, p.units)
		p.units = us
	}
	for _, u := range unit {
		if u.id == 0 {
			continue
		}
		p.units[u.id] = u
	}
}

func (p *Process) UseErrUnit(unit *Unit) {
	if unit.id != 0 {
		panic("invalid error handler unit")
	}
	if p.units == nil {
		p.units = make([]*Unit, 1)
	}
	p.units[0] = unit
}

func (p *Process) Run() {
	for _, u := range p.units {
		u.p = p
		u.ctx = p.ctx
		for i := uint8(0); i < u.concurrent; i++ {
			go u.runLoop()
		}
	}
}

func (p *Process) Cancel() {
	p.cancel()
}

func NewProcess(ctx context.Context) *Process {
	p := &Process{}
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.UseErrUnit(DefaultErrUnit)
	return p
}
