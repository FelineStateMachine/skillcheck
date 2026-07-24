package source

import "fmt"

type Checkpoint struct {
	Fingerprint     Fingerprint `json:"fingerprint"`
	CommittedOffset int64       `json:"committed_offset"`
	ParserRevision  string      `json:"parser_revision"`
}

func (c Checkpoint) AppendEligible(current Fingerprint, parserRevision string) error {
	if c.ParserRevision != parserRevision {
		return fmt.Errorf("parser revision changed")
	}
	if current.Size < c.CommittedOffset {
		return fmt.Errorf("source truncated")
	}
	if current.Identity != c.Fingerprint.Identity {
		return fmt.Errorf("source replaced")
	}
	if current.PrefixDigest != c.Fingerprint.PrefixDigest || current.WindowDigest != c.Fingerprint.WindowDigest {
		return fmt.Errorf("committed prefix digest mismatch")
	}
	return nil
}
