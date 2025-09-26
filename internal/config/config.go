package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-default:"NOT"`
	LocalPathStoage string `yaml:"local_path_storage"`
	HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	const op = "TaskDownloader.internal.configs.Mustload"
	
	path := fetchConfigPath()
	if path == "" {
		panic("Invalid path. Erorr: " + op)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("Invalid read config. Erorr: " + op)
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	fmt.Println(res)

	if res == "" {
		res = os.Getenv("CONFIG")
	}

	return res
}