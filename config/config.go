package config

type Config struct {
	DestinationFolder string
	ForceHashCheck    bool
	CheckLink         bool
}

func GetConfig() Config {
	return Config{
		DestinationFolder: `c:\ws\`,
		ForceHashCheck:    false,
		CheckLink:         false,
	}
}
