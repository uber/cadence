package postgres

import (
    "fmt"

    _ "github.com/lib/pq" // needed to load the postgres driver

    "github.com/iancoleman/strcase"
    "github.com/jmoiron/sqlx"
    "github.com/uber/cadence/tools/sql"
)

const(
    driverName = "postgres"

    dataSourceNamePostgres = "user=%v password=%v host=%v port=%v dbname=%v sslmode=disable "

    readSchemaVersionPostgres     = `SELECT curr_version from schema_version where db_name=?`

    writeSchemaVersionPostgres       = `INSERT into schema_version(db_name, creation_time, curr_version, min_compatible_version) VALUES (?,?,?,?)
										ON CONFLICT (db_name) DO UPDATE 
										  SET creation_time = excluded.creation_time,
										   	  curr_version = excluded.curr_version,
										      min_compatible_version = excluded.min_compatible_version;`


    writeSchemaUpdateHistoryPostgres = `INSERT into schema_update_history(year, month, update_time, old_version, new_version, manifest_md5, description) VALUES(?,?,?,?,?,?,?)`

    createSchemaVersionTablePostgres = `CREATE TABLE schema_version(db_name VARCHAR(255) not null PRIMARY KEY, ` +
        `creation_time TIMESTAMP, ` +
        `curr_version VARCHAR(64), ` +
        `min_compatible_version VARCHAR(64));`

    createSchemaUpdateHistoryTablePostgres = `CREATE TABLE schema_update_history(` +
        `year int not null, ` +
        `month int not null, ` +
        `update_time TIMESTAMP not null, ` +
        `description VARCHAR(255), ` +
        `manifest_md5 VARCHAR(64), ` +
        `new_version VARCHAR(64), ` +
        `old_version VARCHAR(64), ` +
        `PRIMARY KEY (year, month, update_time));`


    createDatabasePostgres = "CREATE database %v"

    dropDatabasePostgres = "Drop database %v"

    listTablesPostgres = "select table_name from information_schema.tables where table_schema='public'"

    dropTablePostgres = "DROP TABLE %v"
)

func init(){
    sql.SupportedSQLDrivers[driverName] = true
    sql.NewConnectionFuncs[driverName] = newPostgresConn
    sql.ReadSchemaVersionSQL[driverName] = readSchemaVersionPostgres
    sql.WriteSchemaVersionSQL[driverName] = writeSchemaVersionPostgres
    sql.WriteSchemaUpdateHistorySQL[driverName] = writeSchemaUpdateHistoryPostgres
    sql.CreateSchemaVersionTableSQL[driverName] = createSchemaVersionTablePostgres
    sql.CreateSchemaUpdateHistoryTableSQL[driverName] = createSchemaUpdateHistoryTablePostgres
    sql.CreateDatabaseSQL[driverName] = createDatabasePostgres
    sql.DropDatabaseSQL[driverName] = dropDatabasePostgres
    sql.ListTablesSQL[driverName] = listTablesPostgres
    sql.DropTableSQL[driverName] = dropTablePostgres
}

func newPostgresConn(driverName, host string, port int, user string, passwd string, database string) (*sqlx.DB, error) {
    if database == ""{
        database = "postgres"
    }
    db, err := sqlx.Connect(driverName, fmt.Sprintf(dataSourceNamePostgres, user, passwd, host, port, database))

    if err != nil {
        return nil, err
    }
    // Maps struct names in CamelCase to snake without need for db struct tags.
    db.MapperFunc(strcase.ToSnake)
    return db, nil
}