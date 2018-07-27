package store

type Event struct {
	Action  string `json:"action"`
	Topic   string `json:"topic"`
	StartOn int64  `json:"starton"`
	Addr    string `json:"addr"`
	Token   string `json:"token"`
	Domain  string `json:"domain"`
}

type Store interface {
	Online(*Event) error
	Offline(*Event) error
}
