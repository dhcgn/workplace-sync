package config

type Config struct {
	DestinationFolder string
}

func GetConfig() Config {
	return Config{
		DestinationFolder: `c:\ws\`,
	}
}
