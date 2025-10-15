package configs

type AuthConfig struct {
	JWTSecret            string `json:"jwt_secret"`              // HMAC 密钥
	AccessTokenTTLMinute int    `json:"access_token_ttl_minute"` // 访问令牌有效期(分钟)
}

type authConfigRaw struct {
	JWTSecret            *string `json:"jwt_secret"`
	AccessTokenTTLMinute *int    `json:"access_token_ttl_minute"`
}

var DefaultAuthConfig = AuthConfig{
	JWTSecret:            "dev-secret-change-me", // 生产务必覆盖！
	AccessTokenTTLMinute: 120,                    // 2h
}

func mergeAuth(dst *AuthConfig, raw *authConfigRaw) {
	if raw == nil {
		return
	}
	if s := strPtrValid(raw.JWTSecret); s != "" {
		dst.JWTSecret = s
	}
	if v := intPtrPos(raw.AccessTokenTTLMinute); v > 0 && v <= 24*60 {
		dst.AccessTokenTTLMinute = v
	}
}
