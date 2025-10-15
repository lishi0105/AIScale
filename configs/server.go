package configs

// ServerConfig HTTP 服务相关配置
type ServerConfig struct {
	Port    int    `json:"port"`     // 监听端口
	WebRoot string `json:"web_root"` // Web 根路径(可选)
}

type serverConfigRaw struct {
	Port    *int    `json:"port"`
	WebRoot *string `json:"web_root"`
}

// 默认 HTTP 配置
var DefaultServerConfig = ServerConfig{
	Port:    7380,
	WebRoot: "./web", // 若 WebRoot 未指定，默认为 "./web"
}

func mergeServer(dst *ServerConfig, raw *serverConfigRaw) {
	if raw == nil {
		return
	}
	if p := intPtrInRange(raw.Port, 1, 65535); p > 0 {
		dst.Port = p
	}
	if s := strPtrNonEmpty(raw.WebRoot); s != "" {
		dst.WebRoot = s
	}
}
