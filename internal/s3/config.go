package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
	Profile        string `json:"profile" yaml:"profile"`
	AccessKey      string `json:"access_key" yaml:"access_key"`
	SecretKey      string `json:"secret_key" yaml:"secret_key"`
	Region         string `json:"region" yaml:"region"`
	Endpoint       string `json:"endpoint" yaml:"endpoint"`
	DisableSSL     bool   `yaml:"disable_ssl"`
	Bucket         string `json:"bucket" yaml:"bucket"`
	ForcePathStyle bool   `yaml:"force_path_style"`
}

func (c Config) createSession() (*session.Session, error) {
	if c.AccessKey != "" && c.SecretKey != "" {
		cfgs := aws.NewConfig().WithRegion(c.Region)
		cred := credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
		cfgs = cfgs.WithCredentials(cred)
		return session.NewSession(cfgs)
	}

	return session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(c.Region)},
		Profile:           c.Profile,
		SharedConfigState: session.SharedConfigEnable,
	})
}