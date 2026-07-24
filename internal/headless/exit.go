package headless

const (
	ExitOK        = 0
	ExitFailure   = 1
	ExitUsage     = 2
	ExitNotFound  = 3
	ExitConflict  = 4
	ExitBusy      = 5
	ExitCancelled = 130
)

func ExitCode(errorCode string) int {
	switch errorCode {
	case "invalid_arguments", "invalid_policy", "unsupported_format":
		return ExitUsage
	case "not_found", "source_unavailable":
		return ExitNotFound
	case "conflict", "revision_conflict":
		return ExitConflict
	case "writer_busy":
		return ExitBusy
	case "cancelled":
		return ExitCancelled
	default:
		return ExitFailure
	}
}
