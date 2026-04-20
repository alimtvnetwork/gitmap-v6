// Package store — projecttype.go manages the ProjectTypes reference table.
package store

import (
	"github.com/user/gitmap/constants"
)

// SeedProjectTypes inserts all supported project types if not present.
func (db *DB) SeedProjectTypes() error {
	_, err := db.conn.Exec(constants.SQLSeedProjectTypes)

	return err
}
