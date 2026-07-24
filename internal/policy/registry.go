package policy

type Registry struct{ Plan Plan }

func NewRegistry(plan Plan) Registry { return Registry{Plan: plan} }

func (r Registry) Correction(id string) (Correction, bool) {
	for _, item := range r.Plan.Corrections {
		if item.ID == id {
			return item, true
		}
	}
	return Correction{}, false
}
