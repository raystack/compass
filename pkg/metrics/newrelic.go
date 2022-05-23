package metrics

type NewRelicConfig struct {
	Enabled    bool   `mapstructure:"enabled" default:"false"`
	AppName    string `mapstructure:"appname" default:"compass"`
	LicenseKey string `mapstructure:"licensekey" default:""`
}
