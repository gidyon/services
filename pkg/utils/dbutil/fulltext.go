package dbutil

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const indexName = "fts_search_index"

// CreateFullTextIndex creates a full-text index
func CreateFullTextIndex(db *gorm.DB, tableName string, columns ...string) error {
	if ok := db.Migrator().HasIndex(tableName, indexName); ok {
		return nil
	}
	sqlQuery := fmt.Sprintf(
		"CREATE FULLTEXT INDEX %s ON %s(%s)",
		indexName, tableName, strings.Join(columns, ","),
	)
	err := db.Table(tableName).Exec(sqlQuery).Error
	return err
}

// DropFullTextIndex drops a full-text index
func DropFullTextIndex(db *gorm.DB, tableName string) error {
	if ok := db.Migrator().HasIndex(tableName, indexName); !ok {
		return nil
	}
	sqlQuery := fmt.Sprintf("ALTER TABLE %s	DROP INDEX %s", tableName, indexName)
	err := db.Table(tableName).Exec(sqlQuery).Error
	return err
}
