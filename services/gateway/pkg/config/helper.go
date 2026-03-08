package config

// GetGlobalConfig 获取全局配置对象
func GetGlobalConfig() *Config {
	return globalReader.GetConfig()
}

// ReloadConfig 重新加载配置（用于热更新场景）
func ReloadConfig() (*Config, error) {
	newReader := NewConfigReader()
	cfg, err := newReader.Load()
	if err != nil {
		return nil, err
	}

	// 用新的配置替换全局实例
	globalReader = newReader
	return cfg, nil
}

// GetString 从全局配置读取器获取字符串值
func GetString(key string) string {
	return globalReader.GetString(key)
}

// GetInt 从全局配置读取器获取整数值
func GetInt(key string) int {
	return globalReader.GetInt(key)
}

// Get 从全局配置读取器获取配置值
func Get(key string) interface{} {
	return globalReader.Get(key)
}

// ValidateConfig 验证当前配置
func ValidateConfig() error {
	return globalReader.ValidateConfig()
}
