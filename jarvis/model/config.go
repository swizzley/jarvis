package model

import (
	"fmt"
	"os"

	"github.com/blendlabs/spiffy"
)

type ConfigDb struct {
	Server   string
	Schema   string
	User     string
	Password string
	DSN      string
}

func (db *ConfigDb) InitFromEnvironment() {
	db.Server = os.Getenv("DB_HOST")
	db.Schema = os.Getenv("DB_SCHEMA")
	db.User = os.Getenv("DB_USER")
	db.Password = os.Getenv("DB_PASSWORD")
	db.DSN = os.Getenv("DATABASE_URL")
}

func (db *ConfigDb) AsPostgresConnectionString() string {
	if len(db.DSN) != 0 {
		return db.DSN
	} else {
		return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", db.User, db.Password, db.Server, db.Schema)
	}
}

func DbInit() error {
	config := &ConfigDb{}
	config.InitFromEnvironment()
	return SetupDatabaseContext(config)
}

func SetupDatabaseContext(config *ConfigDb) error {
	spiffy.CreateDbAlias("main", spiffy.NewDbConnectionFromDSN(config.AsPostgresConnectionString()))
	spiffy.SetDefaultAlias("main")

	_, dbError := spiffy.DefaultDb().Open()
	if dbError != nil {
		return dbError
	}

	spiffy.DefaultDb().Connection.SetMaxIdleConns(50)
	return nil
}
