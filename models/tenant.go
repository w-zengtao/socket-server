package models

import "time"

// Tenant - model which will holds tenant's basic info
type Tenant struct {
	ID       uint      `gorm:"primary_key" json:"id"`
	ExpireAt time.Time `gorm:"column:expire_time" json:"expire_at"`
	APIKey   string    `gorm:"column:api_key;unique_index" sql:"not null" json:"api_key"`
}

// TableName set table's name
func (Tenant) TableName() string {
	return "tenants"
}

// Verify 校验链接是否有权限、目前的业务
// 1. 能找到商户
// 2. APIKey 没有过期
func (tenant *Tenant) Verify(key string) bool {
	if err := tenant.recordByKey(key); err != nil {
		return false
	}
	return tenant.valid()
}

// recordByKey 查找商户
func (tenant *Tenant) recordByKey(key string) error {
	if result := DB.Where(&Tenant{APIKey: key}).First(&tenant); result.Error != nil {
		return result.Error
	}
	return nil
}

// expired 判断 tenant 是否 valid?
func (tenant *Tenant) valid() bool {
	if tenant.ExpireAt.Unix() < time.Now().Unix() {
		return false
	}
	return true
}
