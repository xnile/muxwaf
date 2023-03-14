package repository

import (
	"gorm.io/gorm"
)

type Repository struct {
	DB               *gorm.DB
	User             UserRepo
	BlacklistIP      IBlacklistIPRepo
	Cert             ICertRepo
	Site             ISiteRepo
	SiteConfigRepo   ISiteConfigRepo
	SiteOriginRepo   ISiteOriginRepo
	WhitelistIPRepo  IWhitelistIPRepo
	WhitelistURLRepo IWhitelistURLRepo
	RateLimitRepo    IRateLimitRepo
}

func New(db *gorm.DB) *Repository {
	return &Repository{
		DB:               db,
		User:             NewUserRepo(db),
		BlacklistIP:      NewBlacklistIPRepo(db),
		Cert:             NewCertRepo(db),
		Site:             NewSiteRepo(db),
		SiteConfigRepo:   NewSiteConfigRepo(db),
		SiteOriginRepo:   NewSiteOriginRepo(db),
		WhitelistIPRepo:  NewWhitelistIPRepo(db),
		WhitelistURLRepo: NewWhitelistURLRepo(db),
		RateLimitRepo:    NewRateLimitRepo(db),
	}
}
