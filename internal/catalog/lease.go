package catalog

import (
	"context"
	"errors"
	"time"
)

var ErrWriterBusy = errors.New("catalog writer is busy")

func (c *Catalog) AcquireWriter(ctx context.Context, owner string, ttl time.Duration) (func() error, error) {
	now := time.Now().UnixNano()
	expires := now + ttl.Nanoseconds()
	res, err := c.db.ExecContext(ctx, `INSERT INTO writer_lease(singleton,owner,expires_at) VALUES(1,?,?) ON CONFLICT(singleton) DO UPDATE SET owner=excluded.owner,expires_at=excluded.expires_at WHERE writer_lease.expires_at<? OR writer_lease.owner=?`, owner, expires, now, owner)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrWriterBusy
	}
	return func() error {
		_, err := c.db.Exec(`DELETE FROM writer_lease WHERE singleton=1 AND owner=?`, owner)
		return err
	}, nil
}
