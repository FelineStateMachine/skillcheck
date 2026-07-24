package report

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skilltrace/internal/report/generated"
)

//go:embed assets/report.js assets/report.css
var assets embed.FS

func Export(path string, dataset generated.Dataset) error {
	data, err := json.Marshal(dataset)
	if err != nil {
		return err
	}
	data = bytes.ReplaceAll(data, []byte("<"), []byte("\\u003c"))
	data = bytes.ReplaceAll(data, []byte(">"), []byte("\\u003e"))
	data = bytes.ReplaceAll(data, []byte("&"), []byte("\\u0026"))
	css, err := assets.ReadFile("assets/report.css")
	if err != nil {
		return err
	}
	js, err := assets.ReadFile("assets/report.js")
	if err != nil {
		return err
	}
	html := fmt.Sprintf("<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>Skill usage report</title><style>%s</style></head><body><header><h1 id=\"title\">Skill usage report</h1><p id=\"metadata\"></p></header><main><div id=\"comparison\"></div><h2>Findings and differences</h2><ul id=\"findings\"></ul><noscript>This report requires JavaScript for interaction. Its data remains embedded in this file.</noscript></main><script id=\"report-data\" type=\"application/json\">%s</script><script>%s</script></body></html>", css, data, js)
	if strings.Contains(strings.ToLower(html), "https://") || strings.Contains(strings.ToLower(html), "http://") {
		return fmt.Errorf("report contains a network reference")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".skilltrace-report-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(html); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
