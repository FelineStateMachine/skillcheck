package policy

import "fmt"

func (m Model) View() string {
	if m.Err != "" {
		return "POLICY CONFLICT\n" + m.Err
	}
	if m.Applied {
		return "POLICY APPLIED\nRevision " + m.Preview.Revision
	}
	if m.Preview.Token == "" {
		return "POLICY\nValidate or preview a policy file."
	}
	return fmt.Sprintf("POLICY PREVIEW\n%d corrections, %d findings\nAffected uses: %d\nApply token: %s", m.Preview.Corrections, m.Preview.Findings, m.Preview.AffectedCandidates, m.Preview.Token)
}
