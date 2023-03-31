package utility

import (
	"github.com/BurntSushi/toml"
	"github.com/murasakiakari/pathlib"
)

type Domain struct {
	Name string
}

type TLS struct {
	Port     string
	CertFile string
	KeyFile  string
}

type Server struct {
	Mode string
}

type Backend struct {
	Address string
}

type Log struct {
	FileName string
}

type Configuration struct {
	Domain  *Domain
	TLS     *TLS
	Server  *Server
	Backend *Backend
	Log     *Log
}

var Config Configuration

func init() {
	configurationPath := pathlib.CurrentExecutablePath.Dir().Join("configuration.toml")
	configurationData, err := configurationPath.ReadFile()
	if err != nil {
		panic(err)
	}

	_, err = toml.Decode(string(configurationData), &Config)
	if err != nil {
		panic(err)
	}
}
