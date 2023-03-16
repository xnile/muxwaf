package service

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/xnile/muxwaf/internal/event"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/repository"
	xlog "github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/orm"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var SVC *Service

type Service struct {
	User        UserService
	BlacklistIP IBlacklistIPService
	Cert        ICertService
	Site        ISiteService
	Whitelist   IWhitelistService
	RateLimit   IRateLimitService
	AttackLog   IAttackLogService
	Node        INodeService
}

func newService(gDB *gorm.DB, repo *repository.Repository, eventBus *event.EventBus) *Service {
	return &Service{
		User:        NewUserService(gDB, repo),
		BlacklistIP: NewBlacklistIPService(gDB, repo, eventBus),
		Cert:        NewCertService(repo, eventBus),
		Site:        NewSiteService(repo, eventBus),
		Whitelist:   NewWhitelistService(gDB, repo, eventBus),
		RateLimit:   NewRateLimitService(repo, eventBus),
		AttackLog:   NewAttackLogService(gDB),
		Node:        NewINodeService(gDB, eventBus),
	}
}

func Setup() {
	dbHost := viper.GetString("postgresql.host")
	dbUser := viper.GetString("postgresql.username")
	dbPassword := viper.GetString("postgresql.password")
	dbName := viper.GetString("postgresql.db")
	dbPort := viper.GetString("postgresql.port")
	dbDebug := viper.GetBool("postgresql.debug")

	dbLogLeve := logger.Error
	if dbDebug {
		dbLogLeve = logger.Info
	}

	logger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      dbLogLeve,   // Log level
			Colorful:      false,       // 禁用彩色打印
		},
	)
	// mysql
	//dsn := "root:@tcp(127.0.0.1:3306)/muxwaf?charset=utf8mb4&parseTime=true&loc=Local"

	// postgresql
	//dsn := "host=localhost user=muxwaf password=muxwaf dbname=muxwaf port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", dbHost, dbUser, dbPassword, dbName, dbPort)

	ormConfig := orm.Config{
		DSN:         dsn,
		Active:      0,
		Idle:        0,
		IdleTimeout: 0,
		Logger:      logger,
	}

	//db := orm.NewMySQL(&ormConfig)
	db := orm.NewPgSQL(&ormConfig)

	// 创建表
	if err := model.AutoMigrate(db); err != nil {
		xlog.Fatalf("auto migration error: %v", err)
	}

	eventBus := event.NewEventBus(10)
	eventBus.RegisterHandler(event.NewDefaultHandler(db))
	eventBus.StartWorkers(2)

	// Redis初始化
	//redisApp, err := redis.NewRedis()
	//if err != nil {
	//	xlog.Fatal("init redis error: %v\n", err)
	//	debug.PrintStack()
	//	panic(err)
	//}
	//cacheApp := cache.New(redisApp)

	repo := repository.New(db)
	SVC = newService(db, repo, eventBus)
}
