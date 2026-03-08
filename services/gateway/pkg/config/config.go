package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	AppName  string         `mapstructure:"APP_NAME"`
	AppEnv   string         `mapstructure:"APP_ENV"`
	Port     string         `mapstructure:"PORT"`
	Database DatabaseConfig `mapstructure:"DATABASE"`
	Redis    RedisConfig    `mapstructure:"REDIS"`
	Cache    CacheConfig    `mapstructure:"CACHE"`
	Storage  StorageConfig  `mapstructure:"STORAGE"`
	Mail     MailConfig     `mapstructure:"MAIL"`
	Logger   LoggerConfig   `mapstructure:"LOGGER"`
	JWT      JWTConfig      `mapstructure:"JWT"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host        string `mapstructure:"HOST"`
	Port        string `mapstructure:"PORT"`
	User        string `mapstructure:"USER"`
	Password    string `mapstructure:"PASSWORD"`
	Name        string `mapstructure:"NAME"`
	MaxConn     int    `mapstructure:"MAX_CONN"`
	MaxIdleConn int    `mapstructure:"MAX_IDLE_CONN"`
	MaxLifetime int    `mapstructure:"MAX_LIFETIME"`
	AutoMigrate bool   `mapstructure:"AUTO_MIGRATE"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     string `mapstructure:"PORT"`
	Password string `mapstructure:"PASSWORD"`
	DB       int    `mapstructure:"DB"`
	MaxRetry int    `mapstructure:"MAX_RETRY"`
}

type CacheConfig struct {
	Backend     string `mapstructure:"BACKEND"`
	FilePath    string `mapstructure:"FILE_PATH"`
	RedisPrefix string `mapstructure:"REDIS_PREFIX"`
}

type StorageConfig struct {
	Provider        string `mapstructure:"PROVIDER"`
	Region          string `mapstructure:"REGION"`
	Endpoint        string `mapstructure:"ENDPOINT"`
	Bucket          string `mapstructure:"BUCKET"`
	AccessKeyID     string `mapstructure:"ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"SECRET_ACCESS_KEY"`
	SessionToken    string `mapstructure:"SESSION_TOKEN"`
	UsePathStyle    bool   `mapstructure:"USE_PATH_STYLE"`
	DisableSSL      bool   `mapstructure:"DISABLE_SSL"`
}

type MailConfig struct {
	Enabled     bool   `mapstructure:"ENABLED"`
	Provider    string `mapstructure:"PROVIDER"`
	Host        string `mapstructure:"HOST"`
	Port        int    `mapstructure:"PORT"`
	TLSMode     string `mapstructure:"TLS_MODE"`
	Username    string `mapstructure:"USERNAME"`
	Password    string `mapstructure:"PASSWORD"`
	FromName    string `mapstructure:"FROM_NAME"`
	FromMail    string `mapstructure:"FROM_MAIL"`
	TemplateDir string `mapstructure:"TEMPLATE_DIR"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Dir        string `mapstructure:"DIR"`
	Filename   string `mapstructure:"FILENAME"`
	Level      string `mapstructure:"LEVEL"`
	MaxSizeMB  int    `mapstructure:"MAX_SIZE_MB"`
	MaxBackups int    `mapstructure:"MAX_BACKUPS"`
	MaxAgeDays int    `mapstructure:"MAX_AGE_DAYS"`
	Compress   bool   `mapstructure:"COMPRESS"`
}

type JWTConfig struct {
	SecretKey      string `mapstructure:"SECRET_KEY"`
	ExpireHours    int    `mapstructure:"EXPIRE_HOURS"`
	RefreshHours   int    `mapstructure:"REFRESH_HOURS"`
	Issuer         string `mapstructure:"ISSUER"`
	TokenType      string `mapstructure:"TOKEN_TYPE"`
	HeaderName     string `mapstructure:"HEADER_NAME"`
	ContextUserKey string `mapstructure:"CONTEXT_USER_KEY"`
}

// ConfigReader 配置读取器接口
type ConfigReader interface {
	Load() (*Config, error)
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
}

// ViperConfigReader viper 配置读取器实现
type ViperConfigReader struct {
	viper *viper.Viper
	cfg   *Config
}

// NewConfigReader 创建一个新的配置读取器
func NewConfigReader() *ViperConfigReader {
	return &ViperConfigReader{
		viper: viper.New(),
		cfg:   &Config{},
	}
}

// Load 加载配置
func (vcr *ViperConfigReader) Load() (*Config, error) {
	v := vcr.viper

	// 设置默认值
	vcr.setDefaults()

	// 从环境变量读取
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 确定环境，然后加载配置文件
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	log.Println(env)

	if err := vcr.readConfigByEnv(env); err != nil {
		log.Printf("warning: 未找到配置文件: %v", err)
	}

	// 解析到结构体
	if err := v.Unmarshal(vcr.cfg); err != nil {
		return nil, fmt.Errorf("配置解析失败: %w", err)
	}

	// 应用环境变量覆盖
	vcr.applyEnvOverrides()

	return vcr.cfg, nil
}

// setDefaults 设置默认配置值
func (vcr *ViperConfigReader) setDefaults() {
	v := vcr.viper

	// 应用默认值
	v.SetDefault("APP_NAME", "gateway-service")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("PORT", "8080")

	// 数据库默认值
	v.SetDefault("DATABASE.HOST", "localhost")
	v.SetDefault("DATABASE.PORT", "5432")
	v.SetDefault("DATABASE.USER", "postgres")
	v.SetDefault("DATABASE.PASSWORD", "")
	v.SetDefault("DATABASE.NAME", "deepwrite")
	v.SetDefault("DATABASE.MAX_CONN", 25)
	v.SetDefault("DATABASE.MAX_IDLE_CONN", 5)
	v.SetDefault("DATABASE.MAX_LIFETIME", 300)
	v.SetDefault("DATABASE.AUTO_MIGRATE", false)

	// Redis 默认值
	v.SetDefault("REDIS.HOST", "localhost")
	v.SetDefault("REDIS.PORT", "6379")
	v.SetDefault("REDIS.PASSWORD", "")
	v.SetDefault("REDIS.DB", 0)
	v.SetDefault("REDIS.MAX_RETRY", 3)

	// Cache 默认值
	v.SetDefault("CACHE.BACKEND", "redis")
	v.SetDefault("CACHE.FILE_PATH", "./tmp/cache/cache.json")
	v.SetDefault("CACHE.REDIS_PREFIX", "gateway:")

	// Storage 默认值
	v.SetDefault("STORAGE.PROVIDER", "aws-s3")
	v.SetDefault("STORAGE.REGION", "us-east-1")
	v.SetDefault("STORAGE.ENDPOINT", "")
	v.SetDefault("STORAGE.BUCKET", "")
	v.SetDefault("STORAGE.ACCESS_KEY_ID", "")
	v.SetDefault("STORAGE.SECRET_ACCESS_KEY", "")
	v.SetDefault("STORAGE.SESSION_TOKEN", "")
	v.SetDefault("STORAGE.USE_PATH_STYLE", false)
	v.SetDefault("STORAGE.DISABLE_SSL", false)

	// Mail 默认值
	v.SetDefault("MAIL.ENABLED", false)
	v.SetDefault("MAIL.PROVIDER", "smtp")
	v.SetDefault("MAIL.HOST", "localhost")
	v.SetDefault("MAIL.PORT", 25)
	v.SetDefault("MAIL.TLS_MODE", "auto")
	v.SetDefault("MAIL.USERNAME", "")
	v.SetDefault("MAIL.PASSWORD", "")
	v.SetDefault("MAIL.FROM_NAME", "DeepWrite")
	v.SetDefault("MAIL.FROM_MAIL", "no-reply@deepwrite.local")
	v.SetDefault("MAIL.TEMPLATE_DIR", "templates/mail")

	// Logger 默认值
	v.SetDefault("LOGGER.DIR", "logs")
	v.SetDefault("LOGGER.FILENAME", "gateway")
	v.SetDefault("LOGGER.LEVEL", "info")
	v.SetDefault("LOGGER.MAX_SIZE_MB", 100)
	v.SetDefault("LOGGER.MAX_BACKUPS", 30)
	v.SetDefault("LOGGER.MAX_AGE_DAYS", 30)
	v.SetDefault("LOGGER.COMPRESS", true)

	// JWT 默认值
	v.SetDefault("JWT.SECRET_KEY", "deepwrite-dev-jwt-secret")
	v.SetDefault("JWT.EXPIRE_HOURS", 24)
	v.SetDefault("JWT.REFRESH_HOURS", 168)
	v.SetDefault("JWT.ISSUER", "gateway-service")
	v.SetDefault("JWT.TOKEN_TYPE", "Bearer")
	v.SetDefault("JWT.HEADER_NAME", "Authorization")
	v.SetDefault("JWT.CONTEXT_USER_KEY", "current_user")

	// Workspace 默认值
	v.SetDefault("WORKSPACE.INVITATION_EXPIRE_HOURS", 168)
}

func (vcr *ViperConfigReader) readConfigByEnv(env string) error {
	configPaths := []string{
		".",
		"./config",
		"./pkg/config",
		"../../",
		"../../config",
	}

	candidateFiles := []string{
		fmt.Sprintf("config.%s.yaml", env),
		fmt.Sprintf("config.%s.yml", env),
		"config.yaml",
		"config.yml",
	}

	checkedPaths := make([]string, 0, len(configPaths)*len(candidateFiles))

	for _, path := range configPaths {
		for _, file := range candidateFiles {
			candidate := filepath.Join(path, file)
			absPath, err := filepath.Abs(candidate)
			if err == nil {
				checkedPaths = append(checkedPaths, absPath)
			} else {
				checkedPaths = append(checkedPaths, candidate)
			}

			if _, err := os.Stat(candidate); err != nil {
				continue
			}

			vcr.viper.SetConfigFile(candidate)
			if err := vcr.viper.ReadInConfig(); err != nil {
				return fmt.Errorf("读取配置文件失败(%s): %w", candidate, err)
			}

			log.Printf("using config file: %s", vcr.viper.ConfigFileUsed())
			return nil
		}
	}

	return fmt.Errorf("已检查路径: %s", strings.Join(checkedPaths, ", "))
}

// applyEnvOverrides 应用环境变量覆盖
func (vcr *ViperConfigReader) applyEnvOverrides() {
	if env := os.Getenv("APP_NAME"); env != "" {
		vcr.cfg.AppName = env
	}
	if env := os.Getenv("APP_ENV"); env != "" {
		vcr.cfg.AppEnv = env
	}
	if env := os.Getenv("PORT"); env != "" {
		vcr.cfg.Port = env
	}

	// 数据库环境变量覆盖
	if env := os.Getenv("DB_HOST"); env != "" {
		vcr.cfg.Database.Host = env
	}
	if env := os.Getenv("DB_PORT"); env != "" {
		vcr.cfg.Database.Port = env
	}
	if env := os.Getenv("DB_USER"); env != "" {
		vcr.cfg.Database.User = env
	}
	if env := os.Getenv("DB_PASSWORD"); env != "" {
		vcr.cfg.Database.Password = env
	}
	if env := os.Getenv("DB_NAME"); env != "" {
		vcr.cfg.Database.Name = env
	}

	// Redis 环境变量覆盖
	if env := os.Getenv("REDIS_HOST"); env != "" {
		vcr.cfg.Redis.Host = env
	}
	if env := os.Getenv("REDIS_PORT"); env != "" {
		vcr.cfg.Redis.Port = env
	}
	if env := os.Getenv("REDIS_PASSWORD"); env != "" {
		vcr.cfg.Redis.Password = env
	}

	// Cache 环境变量覆盖
	if env := os.Getenv("CACHE_BACKEND"); env != "" {
		vcr.cfg.Cache.Backend = env
	}
	if env := os.Getenv("CACHE_FILE_PATH"); env != "" {
		vcr.cfg.Cache.FilePath = env
	}
	if env := os.Getenv("CACHE_REDIS_PREFIX"); env != "" {
		vcr.cfg.Cache.RedisPrefix = env
	}

	// Storage 环境变量覆盖
	if env := os.Getenv("STORAGE_PROVIDER"); env != "" {
		vcr.cfg.Storage.Provider = env
	}
	if env := os.Getenv("STORAGE_REGION"); env != "" {
		vcr.cfg.Storage.Region = env
	}
	if env := os.Getenv("STORAGE_ENDPOINT"); env != "" {
		vcr.cfg.Storage.Endpoint = env
	}
	if env := os.Getenv("STORAGE_BUCKET"); env != "" {
		vcr.cfg.Storage.Bucket = env
	}
	if env := os.Getenv("STORAGE_ACCESS_KEY_ID"); env != "" {
		vcr.cfg.Storage.AccessKeyID = env
	}
	if env := os.Getenv("STORAGE_SECRET_ACCESS_KEY"); env != "" {
		vcr.cfg.Storage.SecretAccessKey = env
	}
	if env := os.Getenv("STORAGE_SESSION_TOKEN"); env != "" {
		vcr.cfg.Storage.SessionToken = env
	}

	// Mail 环境变量覆盖
	if env := os.Getenv("MAIL_PROVIDER"); env != "" {
		vcr.cfg.Mail.Provider = env
	}
	if env := os.Getenv("MAIL_HOST"); env != "" {
		vcr.cfg.Mail.Host = env
	}
	if env := os.Getenv("MAIL_PORT"); env != "" {
		vcr.cfg.Mail.Port = vcr.viper.GetInt("MAIL_PORT")
	}
	if env := os.Getenv("MAIL_TLS_MODE"); env != "" {
		vcr.cfg.Mail.TLSMode = env
	}
	if env := os.Getenv("MAIL_USERNAME"); env != "" {
		vcr.cfg.Mail.Username = env
	}
	if env := os.Getenv("MAIL_PASSWORD"); env != "" {
		vcr.cfg.Mail.Password = env
	}
	if env := os.Getenv("MAIL_FROM_NAME"); env != "" {
		vcr.cfg.Mail.FromName = env
	}
	if env := os.Getenv("MAIL_FROM_MAIL"); env != "" {
		vcr.cfg.Mail.FromMail = env
	}
	if env := os.Getenv("MAIL_TEMPLATE_DIR"); env != "" {
		vcr.cfg.Mail.TemplateDir = env
	}
	if env := os.Getenv("MAIL_ENABLED"); env != "" {
		vcr.cfg.Mail.Enabled = vcr.viper.GetBool("MAIL_ENABLED")
	}

	// JWT 环境变量覆盖
	if env := os.Getenv("JWT_SECRET_KEY"); env != "" {
		vcr.cfg.JWT.SecretKey = env
	}
	if env := os.Getenv("JWT_ISSUER"); env != "" {
		vcr.cfg.JWT.Issuer = env
	}
}

// Get 获取配置值（通用）
func (vcr *ViperConfigReader) Get(key string) interface{} {
	return vcr.viper.Get(key)
}

// GetString 获取字符串配置值
func (vcr *ViperConfigReader) GetString(key string) string {
	return vcr.viper.GetString(key)
}

// GetInt 获取整数配置值
func (vcr *ViperConfigReader) GetInt(key string) int {
	return vcr.viper.GetInt(key)
}

// GetConfig 获取已加载的配置对象
func (vcr *ViperConfigReader) GetConfig() *Config {
	return vcr.cfg
}

// ValidateConfig 验证配置的有效性
func (vcr *ViperConfigReader) ValidateConfig() error {
	if vcr.cfg.Port == "" {
		return fmt.Errorf("PORT 配置不能为空")
	}
	if vcr.cfg.Database.Host == "" {
		return fmt.Errorf("DATABASE.HOST 配置不能为空")
	}
	if vcr.cfg.Redis.Host == "" {
		return fmt.Errorf("REDIS.HOST 配置不能为空")
	}
	return nil
}

// LoadConfigFile 从指定文件加载配置
func (vcr *ViperConfigReader) LoadConfigFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("配置文件路径无效: %w", err)
	}

	dir := filepath.Dir(absPath)
	file := filepath.Base(absPath)
	nameWithoutExt := strings.TrimSuffix(file, filepath.Ext(file))

	vcr.viper.AddConfigPath(dir)
	vcr.viper.SetConfigName(nameWithoutExt)

	if err := vcr.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	return nil
}

// 全局配置读取器实例
var globalReader *ViperConfigReader

// init 初始化全局配置读取器（自动加载）
func init() {
	globalReader = NewConfigReader()
	if _, err := globalReader.Load(); err != nil {
		log.Printf("warning: 配置加载出现警告: %v", err)
	}
}
