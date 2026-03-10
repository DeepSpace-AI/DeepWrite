package bootstrap

import (
	"fmt"
	"time"

	"github.com/deepwrite/serivces/gateway/models/document"
	"github.com/deepwrite/serivces/gateway/models/user"
	"github.com/deepwrite/serivces/gateway/models/workspace"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	gatewaylogger "github.com/deepwrite/serivces/gateway/pkg/logger"
	"gorm.io/gorm/logger"
)

func SetupDB() {
	cfg := config.GetGlobalConfig()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	if err := database.Connect(dsn, gatewaylogger.NewGormLogger(logger.Info)); err != nil {
		panic(err)
	}

	// 设置数据库连接池参数
	database.SQLDB.SetMaxOpenConns(cfg.Database.MaxConn)                                     // 设置最大连接数
	database.SQLDB.SetMaxIdleConns(cfg.Database.MaxIdleConn)                                 // 设置最大空闲连接数
	database.SQLDB.SetConnMaxLifetime(time.Duration(cfg.Database.MaxLifetime) * time.Second) // 设置连接最大生命周期

	if cfg.Database.AutoMigrate {
		AutoMigrate()
	}
}

func AutoMigrate() {
	if err := database.DB.AutoMigrate(
		&user.User{},
		&user.Profile{},
		&workspace.Workspace{},
		&workspace.Members{},
		&workspace.Folder{},
		&workspace.WorkspaceFile{},
		&workspace.WorkspaceInvitation{},
		&document.Document{},
		&document.Version{},
		&document.CollabToken{},
		&document.CollabUpdate{},
		&document.CollabState{},
		&document.CollabAudit{},
	); err != nil {
		panic(err)
	}

	if err := database.DB.Exec(`ALTER TABLE documents ALTER COLUMN folder_id DROP NOT NULL`).Error; err != nil {
		panic(err)
	}
}
