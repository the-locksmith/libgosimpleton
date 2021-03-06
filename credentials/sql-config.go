package libgocredentials

import(
	sqldb "database/sql"
)


type SqlConfig struct {
	Base, Table, columnUser, columnHash, columnHashType, columnSalt string
	fullTable, allColumns string
	user User
	Database *sqldb.DB
}

func (config *SqlConfig) init() {
	if config.columnUser == "" {
		config.columnUser = "user"
	}

	if config.columnHash == "" {
		config.columnHash = "password"
	}

	//config.fullTable = config.base + "." + config.table
	config.allColumns = config.columnUser + ", " +
				config.columnHash

	if config.columnHashType != "" {
		config.allColumns += ", " + config.columnHashType
	}

	if config.columnSalt != "" {
		config.allColumns += ", " + config.columnSalt
	}

}
