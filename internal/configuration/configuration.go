package configuration

import (
	"github.com/jinzhu/configor"
)

// testMode indicates if the package is run in test mode
var testMode bool

// Argon2Config holds the parameters used for creating hashes with Argon2.
type Argon2Config struct {
	Memory      uint32 `default:"131072"`
	Iterations  uint32 `default:"4"`
	Parallelism uint8  `default:"4"`
	SaltLength  uint32 `default:"16"`
	KeyLength   uint32 `default:"32"`
}

// CryptoConfig holds the parameters used for creating hashes.
type CryptoConfig struct {
	Argon2 Argon2Config
}

// Formatting holds additional parameters used for formatting messages
type Formatting struct {
	ColoredTitle bool `default:"false"`
}

// Matrix holds credentials for a matrix account
type Matrix struct {
	Homeserver string `default:"https://matrix.org"`
	Username   string `required:"true"`
	Password   string `required:"true"`
}

// Configuration holds values that can be configured by the user.
type Configuration struct {
	Debug bool `default:"false"`
	HTTP  struct {
		ListenAddress string `default:""`
		Port          int    `default:"8080"`
	}
	Database struct {
		Dialect    string `default:"sqlite3"`
		Connection string `default:"pushbits.db"`
	}
	Admin struct {
		Name     string `default:"admin"`
		Password string `default:"admin"`
		MatrixID string `required:"true"`
	}
	Matrix   Matrix
	Security struct {
		CheckHIBP bool `default:"false"`
	}
	Crypto     CryptoConfig
	Formatting Formatting
}

func configFiles() []string {
	if testMode {
		return []string{"config_unittest.yml"}
	}
	return []string{"config.yml"}
}

// Get returns the configuration extracted from env variables or config file.
func Get() *Configuration {
	config := &Configuration{}

	err := configor.New(&configor.Config{
		Environment:          "production",
		ENVPrefix:            "PUSHBITS",
		ErrorOnUnmatchedKeys: true,
	}).Load(config, configFiles()...)
	if err != nil {
		panic(err)
	}

	return config
}
