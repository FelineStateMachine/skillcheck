package adapters_test

import (
	"testing"

	"skilltrace/internal/adapters"
)

// Sessions are labelled by the folder they ran in, not the path that leads
// there: "~/Developer/lofi" is recorded as "lofi".
func TestProjectNameKeepsOnlyTheFolder(t *testing.T) {
	cases := map[string]string{
		"/Users/dami/Developer/lofi":       "lofi",
		"/Users/dami/Developer/lofi/":      "lofi",
		"/Users/dami/Developer/skillcheck": "skillcheck",
		"/Users/dami":                      "dami",
		"relative/path/project":            "project",
		"project":                          "project",
		"":                                 "",
		".":                                "",
		"/":                                "",
		"   /Users/dami/Developer/lofi  ":  "lofi",
	}
	for in, want := range cases {
		if got := adapters.ProjectName(in); got != want {
			t.Fatalf("ProjectName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestProjectNameNeverRetainsAParent(t *testing.T) {
	for _, in := range []string{"/Users/dami/Developer/lofi", "/home/someone/secret-client/work"} {
		got := adapters.ProjectName(in)
		for _, leaked := range []string{"/", "Users", "dami", "Developer", "home", "someone", "secret-client"} {
			if got == leaked || len(got) > 0 && got != "lofi" && got != "work" {
				t.Fatalf("ProjectName(%q) = %q, which retains more than the folder", in, got)
			}
		}
	}
}
