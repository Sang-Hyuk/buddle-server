package api

import (
	"buddle-server/internal/db"
	"buddle-server/internal/jwt"
	"buddle-server/internal/log"
	"buddle-server/internal/s3"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var c configure

type configure struct {
	Logger     log.Config `yaml:"logger"`
	DB         db.Config  `yaml:"db"`
	FileBucket s3.Config  `yaml:"file_bucket"`
	Jwt        jwt.Jwt    `yaml:"jwt"`
}

func InitConfig(p string) error {
	return UnmarshalConfig(p, &c)
}

func UnmarshalConfig(path string, configObj interface{}) error {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "read config file")
	}
	if err := yaml.Unmarshal(configBytes, configObj); err != nil {
		return errors.Wrap(err, "unmarshal config")
	}

	return nil
}

func Config() configure {
	return c
}
