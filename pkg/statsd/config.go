package statsd

// Config represents configuration options for statsd reporter.
type Config struct {
	Enabled             bool    `mapstructure:"enabled" default:"false"`
	Address             string  `mapstructure:"address" default:"127.0.0.1:8125"`
	Prefix              string  `mapstructure:"prefix" default:"compassApi"`
	SamplingRate        float64 `mapstructure:"sampling_rate" default:"1"`
	Separator           string  `mapstructure:"separator" default:"."`
	WithInfluxTagFormat bool    `mapstructure:"with_influx_tag_format" default:"true"`
}
