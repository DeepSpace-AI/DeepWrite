# 配置读取器（Config Reader）

通用的、高效的 Go 应用配置读取管理工具，支持多种格式和环境。

## 特性

- ✅ **多环境支持** - 自动加载 development、test、production 特定配置
- ✅ **环境变量覆盖** - 环境变量优先级更高，便于容器部署
- ✅ **默认值** - 开箱即用，无需手动指定默认配置
- ✅ **配置验证** - 内置验证方法确保核心配置有效
- ✅ **灵活加载** - 支持从指定文件或目录加载配置
- ✅ **全局访问** - 提供全局函数快速访问配置
- ✅ **热更新** - 支持重新加载配置（用于特定场景）
- ✅ **类型安全** - 强类型配置结构体，避免运行时错误

## 使用方法

### 1. 基础使用（推荐）

```go
package main

import (
	"fmt"
	"github.com/deepwrite/serivces/gateway/pkg/config"
)

func main() {
	// 获取全局配置对象
	cfg := config.GetGlobalConfig()
	
	fmt.Printf("端口: %s\n", cfg.Port)
	fmt.Printf("数据库: %s:%s\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("Redis: %s:%s\n", cfg.Redis.Host, cfg.Redis.Port)
}
```

### 2. 创建读取器实例

```go
reader := config.NewConfigReader()
cfg, err := reader.Load()
if err != nil {
	panic(err)
}

// 验证配置
if err := reader.ValidateConfig(); err != nil {
	panic(err)
}

fmt.Printf("应用配置: %+v\n", cfg)
```

### 3. 加载特定配置文件

```go
customReader := config.NewConfigReader()
customReader.LoadConfigFile("./custom-config.yaml")
cfg, _ := customReader.Load()
```

### 4. 全局便捷函数

```go
// 获取字符串值
port := config.GetString("PORT")

// 获取整数值
redisDB := config.GetInt("REDIS.DB")

// 获取通用值
value := config.Get("DATABASE.HOST")

// 获取整个配置对象
cfg := config.GetGlobalConfig()

// 验证配置
err := config.ValidateConfig()

// 重新加载配置
newCfg, err := config.ReloadConfig()
```

## 配置结构

### 主配置对象

```go
type Config struct {
	AppName  string           // 应用名称
	AppEnv   string           // 环境: development, test, production
	Port     string           // 服务端口
	Database DatabaseConfig   // 数据库配置
	Redis    RedisConfig      // Redis 配置
}
```

### 数据库配置

```go
type DatabaseConfig struct {
	Host    string // 数据库主机
	Port    string // 数据库端口
	User    string // 用户名
	Password string // 密码
	Name    string // 数据库名
	MaxConn int    // 最大连接数
	MinConn int    // 最小连接数
}
```

### Redis 配置

```go
type RedisConfig struct {
	Host     string // Redis 主机
	Port     string // Redis 端口
	Password string // 密码
	DB       int    // 数据库号
	MaxRetry int    // 最大重试次数
}
```

## 默认值

| 配置项 | 默认值 |
|--------|--------|
| APP_NAME | gateway-service |
| APP_ENV | development |
| PORT | 8080 |
| DB_HOST | localhost |
| DB_PORT | 5432 |
| DB_USER | postgres |
| REDIS_HOST | localhost |
| REDIS_PORT | 6379 |
| REDIS_DB | 0 |

## 配置加载顺序

1. 设置默认值
2. 从配置文件加载（按 APP_ENV 加载对应文件）
3. 环境变量覆盖（优先级最高）

### 配置文件名规则

- 开发环境: `config.development.yaml`
- 测试环境: `config.test.yaml`
- 生产环境: `config.production.yaml`
- 通用配置: `config.yaml`（备用）

## 环境变量

所有配置项都可以通过环境变量覆盖：

```bash
export APP_NAME=my-service
export APP_ENV=production
export PORT=9090
export DB_HOST=db.example.com
export DB_PORT=5432
export DB_USER=admin
export DB_PASSWORD=secret
export DB_NAME=mydb
export REDIS_HOST=redis.example.com
export REDIS_PORT=6379
export REDIS_PASSWORD=redispass
export REDIS_DB=1
```

## 示例配置文件

### 开发环境 (config.development.yaml)

```yaml
APP_NAME: gateway-service
APP_ENV: development
PORT: "8080"

DATABASE:
  HOST: localhost
  PORT: "5432"
  USER: postgres
  PASSWORD: postgres
  NAME: deepwrite_dev
  MAX_CONN: 25
  MIN_CONN: 5

REDIS:
  HOST: localhost
  PORT: "6379"
  PASSWORD: ""
  DB: 0
  MAX_RETRY: 3
```

### 生产环境 (config.production.yaml)

```yaml
APP_NAME: gateway-service
APP_ENV: production
PORT: "8080"

DATABASE:
  HOST: ${DB_HOST}
  PORT: "${DB_PORT}"
  USER: ${DB_USER}
  PASSWORD: ${DB_PASSWORD}
  NAME: ${DB_NAME}
  MAX_CONN: 50
  MIN_CONN: 10

REDIS:
  HOST: ${REDIS_HOST}
  PORT: "${REDIS_PORT}"
  PASSWORD: ${REDIS_PASSWORD}
  DB: 0
  MAX_RETRY: 5
```

## 错误处理

```go
reader := config.NewConfigReader()
cfg, err := reader.Load()
if err != nil {
	log.Fatalf("配置加载失败: %v", err)
}

// 验证关键配置
if err := reader.ValidateConfig(); err != nil {
	log.Fatalf("配置验证失败: %v", err)
}
```

## 单元测试

```bash
go test ./pkg/config -v
```

## 集成示例

```go
package main

import (
	"fmt"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.GetGlobalConfig()
	
	// 设置 Gin 模式
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	router := gin.Default()
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"app": cfg.AppName,
			"env": cfg.AppEnv,
		})
	})

	// 启动服务
	addr := fmt.Sprintf(":%s", cfg.Port)
	router.Run(addr)
}
```

## 最佳实践

1. **启动时验证** - 在应用启动时调用 `ValidateConfig()` 验证
2. **使用常量** - 对频繁访问的配置项定义常量
3. **环境隔离** - 不同环境使用不同的配置文件
4. **安全敏感信息** - 生产环境的密码等敏感信息使用环境变量传递
5. **日志记录** - 初始化时记录加载的配置（敏感信息除外）

## 故障排除

### 配置未生效

1. 检查环境变量是否设置正确
2. 确认配置文件在正确的路径
3. 验证 `APP_ENV` 环境变量的值
4. 查看应用初始化日志

### 查看当前配置

```go
cfg := config.GetGlobalConfig()
fmt.Printf("%+v\n", cfg)
```

## 扩展

### 添加新配置项

1. 在 `Config` struct 中添加字段
2. 在 `setDefaults()` 中设置默认值
3. 在 `applyEnvOverrides()` 中添加环境变量覆盖（可选）
4. 在配置文件中添加对应项

```go
type Config struct {
	// ... 现有字段 ...
	NewField string `mapstructure:"NEW_FIELD"`
}

func (vcr *ViperConfigReader) setDefaults() {
	// ... 现有设置 ...
	v.SetDefault("NEW_FIELD", "default-value")
}
```

## License

MIT
