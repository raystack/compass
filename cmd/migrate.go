package cmd

func Migrate() {
	if err := loadConfig(); err != nil {
		log.Fatal(err)
	}

	// TODO: add migration script
}
