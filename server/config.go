package server

import "time"

type Config struct {
	TcpAddr      string                 `mapstructure:"tcp_addr"`
	TcpTLSAddr   string                 `mapstructure:"tcp_tls_addr"`
	HttpAddr     string                 `mapstructure:"http_addr"`
	HttpsAddr    string                 `mapstructure:"https_addr"`
	AdminAddr    string                 `mapstructure:"admin_addr"`
	ConnCheckUrl string                 `mapstructure:"conn_check_url"`
	KeepAlive    time.Duration          `mapstructure:"keepalive"`
	Store        map[string]interface{} `mapstructure:"store"`
}
