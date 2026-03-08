package config

import (
	"fmt"
)

// Example 配置读取器使用示例
func Example() {
	// 方法一：使用全局函数（推荐用于简单场景）
	fmt.Println("===== 方法一：全局函数 =====")
	cfg := GetGlobalConfig()
	fmt.Printf("应用名称: %s\n", cfg.AppName)
	fmt.Printf("应用环境: %s\n", cfg.AppEnv)
	fmt.Printf("服务端口: %s\n", cfg.Port)
	fmt.Printf("数据库: %s:%s\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("Redis: %s:%s\n", cfg.Redis.Host, cfg.Redis.Port)

	// 方法二：创建新的读取器实例（推荐用于复杂场景）
	fmt.Println("\n===== 方法二：创建读取器实例 =====")
	reader := NewConfigReader()
	config, err := reader.Load()
	if err != nil {
		fmt.Printf("加载配置出错: %v\n", err)
		return
	}

	// 验证配置
	if err := reader.ValidateConfig(); err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
		return
	}

	fmt.Printf("应用已成功启动，配置: %+v\n", config)

	// 方法三：从特定文件加载配置
	fmt.Println("\n===== 方法三：从指定文件加载 =====")
	customReader := NewConfigReader()
	if err := customReader.LoadConfigFile("./pkg/config/config.production.yaml"); err != nil {
		fmt.Printf("加载自定义配置出错: %v\n", err)
		return
	}

	customCfg, err := customReader.Load()
	if err != nil {
		fmt.Printf("配置加载出错: %v\n", err)
		return
	}
	fmt.Printf("自定义配置: %+v\n", customCfg)
}
