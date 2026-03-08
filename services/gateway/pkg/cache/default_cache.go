package cache

var defaultCache Cache

func SetDefault(c Cache) {
	defaultCache = c
}

func GetDefault() Cache {
	return defaultCache
}
