package cmd

import "github.com/odpf/columbus/tag/sqlstore"

func Migrate() {
	if err := loadConfig(); err != nil {
		panic(err)
	}

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

	if err := pgClient.AutoMigrate(
		&sqlstore.Template{},
		&sqlstore.Field{},
		&sqlstore.Tag{},
	); err != nil {
		panic(err)
	}
}
