package main

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/lazy-electron-consulting/victron-bluetooth/internal/exporter"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	exporter.Config `mapstructure:",squash"`
	Verbose         bool
}

func readConfig(path string) (cfg Config, err error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	if err = viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("could not read in config file '%s': %w", path, err)
	}

	if err = viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.DecodeHookFuncType(decodeKey)
	}); err != nil {
		return cfg, fmt.Errorf("could not unmarshal config: %w", err)
	}
	return cfg, nil
}

func decodeKey(src, dst reflect.Type, i interface{}) (interface{}, error) {
	if src.Kind() != reflect.String || dst != reflect.SliceOf(reflect.TypeOf(byte(0))) {
		return i, nil
	}

	raw, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("unable to coerce %T to string", i)
	}
	key, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("keys must be a hex string: %w", err)
	}
	return key, nil
}
