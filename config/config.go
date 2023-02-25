package config

type Config struct {
	DestinationFolder string
	ForceHashCheck    bool
}

func GetConfig() Config {
	return Config{
		DestinationFolder: `c:\ws\`,
		ForceHashCheck:    true,
	}
}
