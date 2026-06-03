package networking

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
	"github.com/spf13/viper"
)

//go:embed config.yaml
var embeddedConfig []byte

func setupConfig() alohnetwork.Config {
	if len(embeddedConfig) == 0 {
		panic("embedded config is empty")
	}
	data := []byte(os.ExpandEnv(string(embeddedConfig)))
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	cfg := alohnetwork.Config{}
	if err := v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %w", err))
	}
	return cfg
}