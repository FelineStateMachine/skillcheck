package headless

type PolicyMutationResult struct {
	Updated bool   `json:"updated"`
	File    string `json:"file"`
}
