package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
)

type dbKey struct{}

var (
	WriteDBKey = dbKey{}
	ReadDBKey  = dbKey{}
)

type Config struct {
	Host            string `json:"host" yaml:"host"`
	Port            int    `json:"port" yaml:"port"`
	Database        string `json:"database" yaml:"database"`
	Username        string `json:"username" yaml:"username"`
	Password        string `json:"password" yaml:"password"`
	Verbose         bool   `json:"verbose" yaml:"verbose"`
	MaxOpenConns    int    `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	SSLMode         string `json:"ssl_mode" yaml:"ssl_mode"`
}

func Connect(c Config) (*gorm.DB, error) {
	dsnConfig := mysql.NewConfig()
	dsnConfig.User = c.Username
	dsnConfig.Passwd = c.Password
	dsnConfig.Addr = net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	dsnConfig.DBName = c.Database
	dsnConfig.ParseTime = true
	dsnConfig.InterpolateParams = true
	dsnConfig.Collation = "utf8mb4_unicode_ci"
	dsnConfig.Net = "tcp"
	dsnConfig.Params = map[string]string{
		"charset": "utf8mb4",
	}
	dialect := gorm_mysql.Open(dsnConfig.FormatDSN())

	logrus.Debugf("DSN string : %s", dsnConfig.FormatDSN())

	db, err := gorm.Open(dialect, &gorm.Config{DisableNestedTransaction: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect database")
	}

	if c.Verbose {
		db.Logger = logger.New(log.New(logrus.StandardLogger().Out, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Info,
			Colorful:      true,
		})
	}

	stdDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get standard db object")
	}

	maxOpenConn := 10
	if c.MaxOpenConns > 0 {
		maxOpenConn = c.MaxOpenConns
	}
	maxIdleConn := maxOpenConn / 2
	if c.MaxIdleConns > 0 {
		maxIdleConn = c.MaxIdleConns
	}
	connMaxLifetime := 10 * time.Minute
	if c.ConnMaxLifetime > 0 {
		connMaxLifetime = time.Duration(c.ConnMaxLifetime) * time.Second
	}
	connMaxIdletime := 3 * time.Minute
	if c.ConnMaxIdleTime > 0 {
		connMaxIdletime = time.Duration(c.ConnMaxIdleTime) * time.Second
	}

	stdDB.SetMaxOpenConns(maxOpenConn)
	stdDB.SetConnMaxLifetime(connMaxLifetime)
	stdDB.SetMaxIdleConns(maxIdleConn)
	stdDB.SetConnMaxIdleTime(connMaxIdletime)

	return db, nil
}

func ConnFromContext(c context.Context, key dbKey) (*gorm.DB, error) {
	if c == nil {
		return nil, fmt.Errorf("nil Context")
	}

	v, ok := c.Value(key).(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("nil db")
	}

	return v.WithContext(c), nil
}

func ContextWithConn(c context.Context, key dbKey, db *gorm.DB) context.Context {
	return context.WithValue(c, key, db)
}

func Transaction(c context.Context, fn func(c context.Context) error, opts ...*sql.TxOptions) error {
	conn, err := ConnFromContext(c, WriteDBKey)
	if err != nil {
		return errors.Wrap(err, "failed to get connection")
	}

	tx := conn.Begin(opts...)
	if err := tx.Error; err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err := tx.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
			logrus.Error(err)
		}
	}()

	if err := fn(ContextWithConn(c, WriteDBKey, tx)); err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.Wrap(err, "commit")
	}

	return nil
}
