package cmd

import (
	"fmt"

	"github.com/odpf/columbus/tag/sqlstore"
)

func Migrate() {
	fmt.Println("Preparing migration...")
	if err := loadConfig(); err != nil {
		panic(err)
	}

	fmt.Println("Initiating DB client...")
	pgClient, err := sqlstore.NewPostgreSQLClient(sqlstore.Config{
		Port:     config.DBPort,
		Host:     config.DBHost,
		Name:     config.DBName,
		User:     config.DBUser,
		Password: config.DBPassword,
		SSLMode:  config.DBSSLMode,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Migrating DB...")
	if err := pgClient.AutoMigrate(
		&sqlstore.Template{},
		&sqlstore.Field{},
		&sqlstore.Tag{},
	); err != nil {
		panic(err)
	}

	fmt.Println("Migration done.")
}
