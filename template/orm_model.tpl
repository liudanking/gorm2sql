type {{.StructName}} struct {
	Database   *gorm.DB  `gorm:"-" sql:"-"` // hide this field
	Id         int64     `gorm:"primary_key"`
	UpdateTime time.Time `sql:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	CreateTime time.Time `sql:"default:CURRENT_TIMESTAMP"`
}

func (self *{{.StructName}}) TableName() string {
	// you may change it
	return "{{.TableName}}"
}

func (self *{{.StructName}}) DB() *gorm.DB {
	if self.Database == nil {
		// you may edit it to point to customized db instance
		return db
	} else {
		return self.Database
	}
}

func (self *{{.StructName}}) Get(id int64) error {
	return self.DB().Where("id = ?", id).First(self).Error
}

func (self *{{.StructName}}) Create() error {
	self.CreateTime = time.Now()
	self.UpdateTime = self.CreateTime
	return self.DB().Create(self).Error
}

func (self *{{.StructName}}) Update() error {
	self.UpdateTime = time.Now()
	return self.DB().Save(self).Error
}

func (self *{{.StructName}}) Delete() error {
	return self.DB().Delete(self).Error
}
