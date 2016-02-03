package model

import "github.com/blendlabs/spiffy"

func TableExists(tableName string) bool {
	query := "SELECT 1 FROM	pg_catalog.pg_tables WHERE tablename = $1"
	var value int
	execErr := spiffy.DefaultDb().Query(query, tableName).Scan(&value)
	if execErr != nil {
		return false
	} else {
		return value == 1
	}
}

func ColumnExists(tableName, columName string) bool {
	query := "SELECT 1 FROM information_schema.columns i WHERE i.table_name = $1 and i.column_name = $2"
	var value int
	execErr := spiffy.DefaultDb().Query(query, tableName, columName).Scan(&value)
	if execErr != nil {
		return false
	} else {
		return value == 1
	}
}

func ConstraintExists(tableName, constraintName string) bool {
	query := "select 1 from information_schema.constraint_column_usage where table_name = $1  and constraint_name = $2"
	var value int
	execErr := spiffy.DefaultDb().Query(query, tableName, constraintName).Scan(&value)
	if execErr != nil {
		return false
	} else {
		return value == 1
	}
}

func CreateTableIfNotExists(tableName, createStatement string) error {
	if !TableExists(tableName) {
		execErr := spiffy.DefaultDb().Exec(createStatement)
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func CreateColumnIfNotExists(tableName, columnName, createStatement string) error {
	if !ColumnExists(tableName) {
		execErr := spiffy.DefaultDb().Exec(createStatement)
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func CreateConstraintIfNotExists(tableName, constraintName, alterStatement string) error {
	if !ConstraintExists(tableName, constraintName) {
		execErr := spiffy.DefaultDb().Exec(alterStatement)
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func InitSchema() error {
	var err error

	err = createTableIfNotExists("tracked_stock", "CREATE TABLE tracked_stock(ticker varchar(255) not null, created_by varchar(255), timestamp_utc timestamp not null);")
	if err != nil {
		return err
	}

	err = createConstraintIfNotExists("tracked_stock", "pk_tracked_stock_ticker", "ALTER TABLE tracked_stock ADD CONSTRAINT pk_tracked_stock_ticker PRIMARY KEY (ticker);")
	if err != nil {
		return err
	}
	return nil
}
