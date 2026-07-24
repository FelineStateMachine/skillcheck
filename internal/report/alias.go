package report

import "fmt"

type aliases struct {
	nodes    map[string]string
	variants map[string]string
	models   map[string]string
}

func newAliases() *aliases {
	return &aliases{nodes: map[string]string{}, variants: map[string]string{}, models: map[string]string{}}
}
func (a *aliases) node(value string) string    { return a.value(a.nodes, value, "step") }
func (a *aliases) variant(value string) string { return a.value(a.variants, value, "path") }
func (a *aliases) model(value string) string   { return a.value(a.models, value, "model") }
func (a *aliases) value(values map[string]string, value, prefix string) string {
	if alias, ok := values[value]; ok {
		return alias
	}
	alias := fmt.Sprintf("%s-%d", prefix, len(values)+1)
	values[value] = alias
	return alias
}
