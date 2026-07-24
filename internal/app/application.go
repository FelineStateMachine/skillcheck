package app

import "skilltrace/internal/catalog"

type Application struct{ Catalog *catalog.Catalog }

func New(c *catalog.Catalog) *Application { return &Application{Catalog: c} }
