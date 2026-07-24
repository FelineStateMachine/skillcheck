package app

import (
	"context"
	"errors"
	"time"

	"skilltrace/internal/apperror"
	"skilltrace/internal/catalog"
)

func (a *Application) CatalogStatus() ([]catalog.Health, error) { return a.Catalog.Health() }
func (a *Application) PreviewLifecycle(ctx context.Context, action, target string) (catalog.DeletionPreview, error) {
	return a.Catalog.Preview(ctx, action, target)
}
func (a *Application) ApplyLifecycle(ctx context.Context, preview catalog.DeletionPreview) error {
	release, err := a.Catalog.AcquireWriter(ctx, "lifecycle", 30*time.Second)
	if errors.Is(err, catalog.ErrWriterBusy) {
		return apperror.Wrap("writer_busy", "catalog writer is busy", err)
	}
	if err != nil {
		return err
	}
	defer release()
	return a.Catalog.Execute(ctx, preview)
}
