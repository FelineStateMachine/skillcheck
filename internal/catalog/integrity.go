package catalog

import (
	"fmt"
	"os"
)

func (c *Catalog) IntegrityCheck() error {
	var result string
	if err := c.db.QueryRow(`PRAGMA integrity_check`).Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("catalog integrity check failed")
	}
	if info, err := os.Stat(c.path); err != nil {
		return err
	} else if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("catalog permissions are too broad")
	}
	return nil
}
