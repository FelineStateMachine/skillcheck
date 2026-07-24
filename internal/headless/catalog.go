package headless

import "skilltrace/internal/catalog"

type CatalogLifecycleResult struct {
	Preview catalog.DeletionPreview `json:"preview"`
	Applied bool                    `json:"applied"`
}
