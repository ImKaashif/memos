package db

import (
	"github.com/pkg/errors"

	"github.com/usememos/memos/server/profile"
	"github.com/usememos/memos/store"
	"github.com/usememos/memos/store/db/mysql"
	"github.com/usememos/memos/store/db/sqlite"
)

// NewDBDriver creates new db driver based on profile.
func NewDBDriver(profile *profile.Profile) (store.Driver, error) {
	var driver store.Driver
	var err error

	// As mysql driver is not fully implemented, we use sqlite for now in prod mode.
	if profile.Mode == "prod" && profile.Driver != "sqlite" {
		return nil, errors.New("Only SQLite is supported in prod mode")
	}

	switch profile.Driver {
	case "sqlite":
		driver, err = sqlite.NewDB(profile)
	case "mysql":
		driver, err = mysql.NewDB(profile)
	default:
		return nil, errors.New("unknown db driver")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to create db driver")
	}
	return driver, nil
}
