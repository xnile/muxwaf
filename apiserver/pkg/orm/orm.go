package orm

import (
	"github.com/xnile/muxwaf/pkg/logx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Config struct {
	DSN         string
	Active      int
	Idle        int
	IdleTimeout time.Duration
	Logger      logger.Interface
}

func NewMySQL(c *Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{Logger: c.Logger})
	if err != nil {
		logx.Errorf("db dns(%s) error: %v", c.DSN, err)
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(c.Idle)
	sqlDB.SetMaxOpenConns(c.Active)
	sqlDB.SetConnMaxLifetime(time.Duration(c.IdleTimeout) / time.Second)
	return db
}

func NewPgSQL(c *Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(c.DSN), &gorm.Config{Logger: c.Logger})
	if err != nil {
		logx.Errorf("db dns(%s) error: %v", c.DSN, err)
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(c.Idle)
	sqlDB.SetMaxOpenConns(c.Active)
	sqlDB.SetConnMaxLifetime(time.Duration(c.IdleTimeout) / time.Second)
	return db
}
