package runtime

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const configTag = "config"

// Config stores all of the configuration required for our service
type Config struct {
	HTTPPort        int           `config:"http_port"`
	LogLevel        string        `config:"log_level"`
	ServiceName     string        `config:"service_name"`
	RedisAddress    string        `config:"redis_address"`
	RedisPassword   string        `config:"redis_password"`
	NewFileInterval time.Duration `config:"new_file_interval"`
	MaxWorkers      int           `config:"max_workers"`
}

// NewConfig creates new configs based on different sources and applying them on the config struct
// with the following sequence (last in the list has the highest priority):
//
// 	1) flags
// 	2) config file (e.g. config.json)
// 	3) environment variables
//
func NewConfig(flags *pflag.FlagSet, fileName string, c *Config) error {
	vi := viper.New()
	vi.AutomaticEnv()
	vi.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// extract the filename to use in building paths
	_, file := path.Split(os.Args[0])
	// set up File loading
	vi.SetConfigName(fileName)
	vi.AddConfigPath(fmt.Sprintf("/etc/%s/", file))
	vi.AddConfigPath(fmt.Sprintf("$HOME/.%s", file))
	vi.AddConfigPath(".")
	// bind the command line flags
	if flags != nil {
		if err := vi.BindPFlags(flags); err != nil {
			log.Errorf("failed to bind command line flags: %s", err)
		}
	}
	// read from the config file
	err := vi.ReadInConfig()
	if err == nil && vi.ConfigFileUsed() != "" {
		log.Infof("loaded configuration from: %s", vi.ConfigFileUsed())
	}
	t := reflect.TypeOf(c)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// bind the possible environment variables
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if err = vi.BindEnv(field.Tag.Get(configTag)); err != nil {
			log.Errorf("failed to bind the possible environment variables: %s", err)
		}
	}
	// unmarshal configs into the provided struct
	if err = vi.Unmarshal(&c, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = configTag
	}); err != nil {
		return err
	}
	return nil
}
