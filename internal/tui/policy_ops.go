package tui

import (
	tea "charm.land/bubbletea/v2"

	"skilltrace/internal/app"
	"skilltrace/internal/tui/catalog"
	"skilltrace/internal/tui/policy"
)

type policyOp uint8

const (
	policyValidate policyOp = iota
	policyPreview
	policyApply
)

// policyResult carries a completed policy operation back to the root. Each
// operation updates a different part of the policy model, so the result knows
// how to apply itself rather than the root re-deriving which one ran.
type policyResult struct {
	op         policyOp
	validation app.PolicyValidation
	preview    app.PolicyPreview
}

func (r policyResult) apply(model policy.Model, err error) policy.Model {
	if err != nil {
		return model.Fail(err.Error())
	}
	switch r.op {
	case policyValidate:
		model.Validation = r.validation
		model.Err = ""
		return model
	case policyPreview:
		model.Preview = r.preview
		model.Err = ""
		model.Applied = false
		return model
	default:
		model.Preview = r.preview
		return model.ApplyCompleted()
	}
}

func (m *Root) policyCmd(op policyOp) tea.Cmd {
	id, _ := m.beginOperation()
	file, token := m.config.PolicyFile, m.Policy.Preview.Token
	return func() tea.Msg {
		var (
			result policyResult
			err    error
		)
		result.op = op
		switch op {
		case policyValidate:
			result.validation, err = m.app.ValidatePolicy(file)
		case policyPreview:
			result.preview, err = m.app.PreviewPolicy(file)
		default:
			var applied app.PolicyApplyResult
			applied, err = m.app.ApplyPolicy(file, token)
			result.preview = applied.PolicyPreview
		}
		return OperationFinishedMsg{ID: id, Result: result, Err: err}
	}
}

// applyGate describes the pending destructive write so it can be confirmed
// against the revision-bound token the preview issued. Applying a policy
// rewrites catalog rows, so it is the one action that must not fire on a
// stale preview.
func (m *Root) applyGate() catalog.Model {
	return catalog.Model{
		Action:       "policy apply",
		Target:       m.config.PolicyFile,
		PreviewToken: m.Policy.Preview.Token,
		Affected:     m.Policy.Preview.AffectedCandidates,
		WriterBusy:   m.cancel != nil,
	}
}
