package lib

import (
	"fmt"
	"os"
	"time"

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

func tableExists(tableName string) bool {
	query := "SELECT 1 FROM	pg_catalog.pg_tables WHERE tablename = $1"
	var value int
	execErr := spiffy.DefaultDb().Query(query, tableName).Scan(&value)
	if execErr != nil {
		return false
	} else {
		return value == 1
	}
}

func columnExists(tableName, columName string) bool {
	query := "SELECT 1 FROM information_schema.columns i WHERE i.table_name = $1 and i.column_name = $2"
	var value int
	execErr := spiffy.DefaultDb().Query(query, tableName, columName).Scan(&value)
	if execErr != nil {
		return false
	} else {
		return value == 1
	}
}

func constraintExists(tableName, constraintName string) bool {
	query := "select 1 from information_schema.constraint_column_usage where table_name = $1  and constraint_name = $2"
	var value int
	execErr := spiffy.DefaultDb().Query(query, tableName, constraintName).Scan(&value)
	if execErr != nil {
		return false
	} else {
		return value == 1
	}
}

func createTableIfNotExists(tableName, createStatement string) error {
	if !tableExists(tableName) {
		execErr := spiffy.DefaultDb().Exec(createStatement)
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func createConstraintIfNotExists(tableName, constraintName, alterStatement string) error {
	if !constraintExists(tableName, constraintName) {
		execErr := spiffy.DefaultDb().Exec(alterStatement)
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func MigrateModel() error {
	err := createTableIfNotExists("tracked_stock", "CREATE TABLE tracked_stock(ticker varchar(255) not null, created_by varchar(255), timestamp_utc timestamp not null);")
	if err != nil {
		return err
	}
	err = createConstraintIfNotExists("tracked_stock", "pk_tracked_stock_ticker", "ALTER TABLE tracked_stock ADD CONSTRAINT pk_tracked_stock_ticker PRIMARY KEY (ticker);")
	if err != nil {
		return err
	}
	return nil
}

type TrackedStock struct {
	Ticker       string    `db:"ticker"`
	CreatedBy    string    `db:"created_by"`
	TimestampUTC time.Time `db:"timestamp_utc"`
}

func (ts TrackedStock) TableName() string {
	return "tracked_stock"
}
