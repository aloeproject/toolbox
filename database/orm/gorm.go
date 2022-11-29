package orm

import (
	"fmt"
	"github.com/aloeproject/toolbox/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"log"
	"time"
)

func NewGorm(dataSource string, logger logger.ILogger) *gorm.DB {
	newLogger := NewGormLogger(logger, glogger.Config{
		SlowThreshold: time.Second,  // 慢 SQL 阈值
		LogLevel:      glogger.Info, // Log level
		Colorful:      true,         // 禁用彩色打印
	})

	//"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
	db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{Logger: newLogger})
	if err != nil || db == nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(50)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(150)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}
