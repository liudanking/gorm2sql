package sqlmodel

import (
	"third/gorm"
	"time"
)

type UserBase struct {
	UserId string `sql:"index:idx_ub"`
	Ip     string `sql:"unique_index:uniq_ip"`
}

type UserEmail struct {
	Database *gorm.DB `gorm:"-" sql:"-"` // hide this field
	Id       int64    `gorm:"primary_key"`
	UserBase
	Email      string
	Sex        bool
	Age        int
	Score      float64
	UpdateTime time.Time `sql:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	CreateTime time.Time `sql:"default:CURRENT_TIMESTAMP"`
}
