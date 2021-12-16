package cmd

import (
	"fmt"

	"github.com/odpf/columbus/store/postgres"
)

func Migrate() {
	fmt.Println("Preparing migration...")
	if err := loadConfig(); err != nil {
		panic(err)
	}

	fmt.Println("Initiating DB client...")
	pgClient, err := postgres.NewClient(postgres.Config{
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
		&postgres.Template{},
		&postgres.Field{},
		&postgres.Tag{},
	); err != nil {
		panic(err)
	}

	fmt.Println("Migration done.")
}
