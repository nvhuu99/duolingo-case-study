package mysql

import (
	"context"
	"database/sql"
	"duolingo/lib/migrate"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	connectionTimeOut = 10 * time.Second
	defaultTimeOut    = 30 * time.Second
)

var ()

type MySQL struct {
	ctx context.Context

	conn mysql.Config

	batchNumber int
	timeOut     time.Duration
}

func New(ctx context.Context) *MySQL {
	driver := MySQL{}
	driver.ctx = ctx
	driver.timeOut = defaultTimeOut
	driver.conn = mysql.Config{Net: "tcp"}

	return &driver
}

func (driver *MySQL) SetConnection(host string, port string, usr string, pwd string) {
	driver.conn.User = usr
	driver.conn.Passwd = pwd
	driver.conn.Addr = host + ":" + port
}

func (driver *MySQL) SetDatabase(database string) {
	driver.conn.DBName = database
}

func (driver *MySQL) GetFileExt() string {
	return ".sql"
}

func (driver *MySQL) SetOperationTimeOut(duration time.Duration) {
	driver.timeOut = duration
}

// PrepareDatabase initializes the "migrations" collection and retrieves the last batch number
func (driver *MySQL) PrepareDatabase() error {
	db, err := driver.connect()
	if err != nil {
		return err
	}
	defer db.Close()
	// Create migrations table
	db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		batch_number INT NOT NULL,
		status VARCHAR(20) NOT NULL
	)`)
	// Retrieve the last batch number from the "migrations" collection
	var lastBatchNum int
	row := db.QueryRow(`
		SELECT batch_number FROM migrations 
		WHERE status IN (?)
		ORDER BY batch_number DESC, id DESC
	`, migrate.MigrateFinished)
	if err := row.Scan(&lastBatchNum); err != nil {
		if err == sql.ErrNoRows {
			lastBatchNum = 0
		}
	}
	// Set the next batch number
	driver.batchNumber = lastBatchNum + 1

	return nil
}

func (driver *MySQL) BatchNumber() int {
	return driver.batchNumber
}

// Retrieves the last batch of migrations from the "migrations" collection.
func (driver *MySQL) LastBatch() ([]migrate.Migration, error) {
	db, err := driver.connect()
	if err != nil {
		return []migrate.Migration{}, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT * FROM migrations WHERE batch_number = ? ORDER BY id ASC`,
		driver.batchNumber-1,
	)
	if err != nil {
		return []migrate.Migration{}, err
	}
	defer rows.Close()

	migrations := []migrate.Migration{}
	for rows.Next() {
		migr := migrate.Migration{}
		if err := rows.Scan(&migr.Id, &migr.Name, &migr.BatchNumber, &migr.Status); err != nil {
			return []migrate.Migration{}, err
		}
		migrations = append(migrations, migr)
	}

	return migrations, nil
}

func (driver *MySQL) RunMigration(migr *migrate.Migration) error {
	db, err := driver.connect()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(string(migr.Body))
	if err != nil {
		return err
	}

	return nil
}

func (driver *MySQL) SaveMigrationRecord(migr *migrate.Migration) error {
	db, err := driver.connect()
	if err != nil {
		return err
	}
	defer db.Close()

	stmt := `INSERT INTO migrations(name, batch_number, status) VALUES (?, ?, ?)`
	_, err = db.Exec(stmt, migr.Name, migr.BatchNumber, migr.Status)

	return err
}

func (driver *MySQL) DeleteMigrationRecord(migr *migrate.Migration) error {
	db, err := driver.connect()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM migrations WHERE id = ?`, migr.Id)

	return err
}

func (driver *MySQL) connect() (*sql.DB, error) {
	db, err := sql.Open("mysql", driver.conn.FormatDSN())
	if err != nil {
		return nil, err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		return nil, errors.New("connection failure")
	}

	return db, nil
}
