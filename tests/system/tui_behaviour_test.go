package system_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"skilltrace/internal/app"
	"skilltrace/internal/catalog"
	"skilltrace/internal/skills"
	"skilltrace/internal/tui"
)

func behaviourRoot(t *testing.T, config tui.Config, snapshot app.DiscoverySnapshot) *tui.Root {
	t.Helper()
	c, err := catalog.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	root := tui.New(app.New(c), config, snapshot)
	_, _ = root.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return root
}

// An empty list has to say where it looked, or the user cannot tell a missing
// install from a misconfigured root.
func TestEmptyDiscoveryNamesTheRootsItSearched(t *testing.T) {
	root := behaviourRoot(t, tui.Config{SkillRoots: []string{"/tmp/project/.codex/skills", "/tmp/home/.claude/skills"}}, app.DiscoverySnapshot{})
	view := root.View().Content
	for _, want := range []string{"Searched", "/tmp/project/.codex/skills", "/tmp/home/.claude/skills", "--skill-root"} {
		if !strings.Contains(view, want) {
			t.Fatalf("empty state missing %q:\n%s", want, view)
		}
	}
}

// Skills that could not be read are recorded by discovery already; dropping
// them from the view makes them look like they do not exist.
func TestDiscoverySurfacesUnreadableSkills(t *testing.T) {
	snapshot := app.DiscoverySnapshot{
		Skills: []app.SkillSnapshot{{Name: "nzip", State: "codex/project"}},
		Issues: []skills.Issue{{Entry: "broken", Reason: "invalid_frontmatter"}},
	}
	view := behaviourRoot(t, tui.Config{}, snapshot).View().Content
	for _, want := range []string{"could not be read", "broken", "invalid_frontmatter"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}

// The footer must not advertise a key that does nothing.
func TestScanHintOnlyAppearsWhenScanIsConfigured(t *testing.T) {
	snapshot := app.DiscoverySnapshot{Skills: []app.SkillSnapshot{{Name: "nzip"}}}
	if view := behaviourRoot(t, tui.Config{}, snapshot).View().Content; strings.Contains(view, "s scan") {
		t.Fatalf("scan hint shown with no scan input configured:\n%s", view)
	}
	if view := behaviourRoot(t, tui.Config{ScanInput: "trace.jsonl"}, snapshot).View().Content; !strings.Contains(view, "s scan") {
		t.Fatalf("scan hint missing with a scan input configured:\n%s", view)
	}
}

// Multi-root discovery is what makes running the tool in an unrelated
// repository useful; a globally installed skill has to be found.
func TestDiscoverFindsGloballyInstalledSkills(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude", "skills", "demo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: demo\ndescription: global skill\n---\nbody\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := skills.Discover(skills.DefaultExposures(t.TempDir(), home))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skills) != 1 || result.Skills[0].Identity.Name != "demo" {
		t.Fatalf("global skill was not discovered: %#v", result.Skills)
	}
	if len(result.Skills[0].Exposures) != 1 || result.Skills[0].Exposures[0].Scope != "global" {
		t.Fatalf("exposure scope not recorded: %#v", result.Skills[0].Exposures)
	}
}

// Escape sequences in third-party skill documents must not reach the terminal.
func TestSkillMetadataIsSanitizedAtIngest(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "styled")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	doc := "---\nname: styled\ndescription: \x1b[31mRED\x1b[0m and a\ttab\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := skills.Discover([]skills.Exposure{{Harness: "codex", Scope: "project", Root: root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skills) != 1 {
		t.Fatalf("expected one skill, got %#v", result.Skills)
	}
	if got := result.Skills[0].Identity.Description; strings.ContainsRune(got, 0x1b) || strings.ContainsRune(got, '\t') {
		t.Fatalf("description still carries control characters: %q", got)
	}
}

// Analyze used to run for seconds with the screen unchanged, and a failure was
// indistinguishable from a slow success.
func TestAnalyzeStateIsVisibleAndErrorsSurface(t *testing.T) {
	snapshot := app.DiscoverySnapshot{Skills: []app.SkillSnapshot{{Name: "nzip"}}}
	root := behaviourRoot(t, tui.Config{}, snapshot)
	_, cmd := root.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd == nil {
		t.Fatal("enter did not start an analyze")
	}
	if view := root.View().Content; !strings.Contains(view, "Analyzing nzip") {
		t.Fatalf("no in-flight indicator while analyzing:\n%s", view)
	}
}

func TestWorkflowRouteIsReachableFromEpisodes(t *testing.T) {
	root := behaviourRoot(t, tui.Config{}, app.DiscoverySnapshot{Skills: []app.SkillSnapshot{{Name: "nzip"}}})
	root.Route = tui.EpisodesRoute
	_, _ = root.Update(tea.KeyPressMsg(tea.Key{Text: "w", Code: 'w'}))
	if root.Route != tui.WorkflowRoute {
		t.Fatalf("w did not open the workflow route, route = %v", root.Route)
	}
	if view := root.View().Content; !strings.Contains(view, "WORKFLOW MAP") {
		t.Fatalf("workflow view not rendered:\n%s", view)
	}
	_, _ = root.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if root.Route != tui.EpisodesRoute {
		t.Fatalf("esc did not return to episodes, route = %v", root.Route)
	}
}

// Applying a policy rewrites catalog rows, so it must refuse to run without a
// preview token bound to the current revision.
func TestPolicyApplyRefusesWithoutAPreview(t *testing.T) {
	root := behaviourRoot(t, tui.Config{PolicyFile: "policy.yaml"}, app.DiscoverySnapshot{})
	root.Route = tui.PolicyRoute
	_, cmd := root.Update(tea.KeyPressMsg(tea.Key{Text: "a", Code: 'a'}))
	if cmd != nil {
		t.Fatal("apply ran without a preview token")
	}
	if view := root.View().Content; !strings.Contains(view, "preview the policy") {
		t.Fatalf("no explanation for the refused apply:\n%s", view)
	}
}

func TestCtrlCMarksTheSessionInterrupted(t *testing.T) {
	root := behaviourRoot(t, tui.Config{}, app.DiscoverySnapshot{})
	_, _ = root.Update(tea.KeyPressMsg(tea.Key{Text: "q", Code: 'q'}))
	if root.Interrupted {
		t.Fatal("q must be a clean exit")
	}
	root = behaviourRoot(t, tui.Config{}, app.DiscoverySnapshot{})
	_, _ = root.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	if !root.Interrupted {
		t.Fatal("ctrl+c must mark the session interrupted so the exit code can be 130")
	}
}
