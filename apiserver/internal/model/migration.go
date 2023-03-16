package model

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) (err error) {
	if err = db.AutoMigrate(new(UserModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(SiteModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(BlacklistIPModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(WhitelistIPModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(WhitelistURLModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(RateLimitModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(SiteConfigModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(SiteOriginModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(SiteRegionBlacklistModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(CertModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(AttackLogModel)); err != nil {
		return err
	}
	if err = db.AutoMigrate(new(NodeModel)); err != nil {
		return err
	}
	return
}
