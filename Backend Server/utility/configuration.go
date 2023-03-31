package utility

import (
	"github.com/BurntSushi/toml"
	"github.com/murasakiakari/pathlib"
)

type Database struct {
	Type     string
	Username string
	Password string
	Name     string
}

func (db *Database) OpenConfig() (string, string) {
	return db.Type, db.Username + ":" + db.Password + "@/" + db.Name
}

type Email struct {
	Account  string
	Password string
}

type Ticket struct {
	Offset int
}

type Log struct {
	FileName string
}

type Configuration struct {
	Database *Database
	Email    *Email
	Ticket   *Ticket
	Log      *Log
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
