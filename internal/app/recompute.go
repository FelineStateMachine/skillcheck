package app

import "context"

type Rebuild func(context.Context, string) error

func Recompute(ctx context.Context, keys []string, rebuild Rebuild) error {
	for _, key := range keys {
		if err := rebuild(ctx, key); err != nil {
			return err
		}
	}
	return nil
}
