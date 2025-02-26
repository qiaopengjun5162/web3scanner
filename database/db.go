// Package database provides database-related functionality.
package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/qiaopengjun5162/web3scanner/common/retry"
	"github.com/qiaopengjun5162/web3scanner/config"

	// Register custom serializers for GORM (e.g., U256Serializer, BytesSerializer).
	_ "github.com/qiaopengjun5162/web3scanner/database/utils/serializers"
)

type DB struct {
	gorm      *gorm.DB
	Addresses AddressesDB
}

func NewDB(ctx context.Context, dbConfig config.DBConfig) (*DB, error) {
	dsn := fmt.Sprintf("host=%s dbname=%s sslmode=disable", dbConfig.Host, dbConfig.Name)
	if dbConfig.Port != 0 {
		dsn += fmt.Sprintf(" port=%d", dbConfig.Port)
	}
	if dbConfig.User != "" {
		dsn += fmt.Sprintf(" user=%s", dbConfig.User)
	}
	if dbConfig.Password != "" {
		dsn += fmt.Sprintf(" password=%s", dbConfig.Password)
	}

	gormConfig := gorm.Config{
		SkipDefaultTransaction: true,
		CreateBatchSize:        3_000,
	}

	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	gorm, err := retry.Do[*gorm.DB](context.Background(), 10, retryStrategy, func() (*gorm.DB, error) {
		gorm, err := gorm.Open(postgres.Open(dsn), &gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return gorm, nil
	})

	if err != nil {
		return nil, err
	}

	db := &DB{
		gorm:      gorm,
		Addresses: NewAddressesDB(gorm),
	}
	return db, nil
}

func (db *DB) Transaction(fn func(db *DB) error) error {
	return db.gorm.Transaction(func(tx *gorm.DB) error {
		txDB := &DB{
			gorm:      tx,
			Addresses: NewAddressesDB(tx),
		}
		return fn(txDB)
	})
}

// Close closes the database connection.
//
// It returns an error if closing the connection fails.
func (db *DB) Close() error {
	sql, err := db.gorm.DB()
	if err != nil {
		return err
	}
	return sql.Close()
}

// ExecuteSQLMigration applies all SQL migrations found in the given folder.
//
// It iterates over all files in the folder and executes their content as SQL.
// If any error occurs, it is returned.
func (db *DB) ExecuteSQLMigration(migrationsFolder string) error {
	err := filepath.Walk(migrationsFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to process migration file: %s", path))
		}
		if info.IsDir() {
			return nil
		}

		// Ensure the file is within the migrations folder to prevent path traversal attacks
		relativePath, err := filepath.Rel(migrationsFolder, path)
		if err != nil || strings.Contains(relativePath, "..") {
			return errors.New("invalid migration file path")
		}
		// Read the file content
		fileContent, readErr := os.ReadFile(path)
		if readErr != nil {
			return errors.Wrap(readErr, fmt.Sprintf("Error reading SQL file: %s", path))
		}

		execErr := db.gorm.Exec(string(fileContent)).Error
		if execErr != nil {
			return errors.Wrap(execErr, fmt.Sprintf("Error executing SQL script: %s", path))
		}
		return nil
	})
	return err
}
