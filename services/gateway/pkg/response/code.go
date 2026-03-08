package response

const (
	// 正确代码
	SuccessCode        = 200
	SuccessCreatedCode = 201

	// 错误代码
	ErrorUnknownCode      = 500
	ErrorBadRequestCode   = 400
	ErrorUnauthorizedCode = 401 // 未授权
	ErrorForbiddenCode    = 403 // 禁止访问
)
