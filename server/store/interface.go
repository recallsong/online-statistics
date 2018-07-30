package store

type OnlineEvent struct {
	Action  string `json:"action"`
	Topic   string `json:"topic"`
	StartOn int64  `json:"starton"`
	Addr    string `json:"addr"`
	Token   string `json:"token"`
	Domain  string `json:"domain"`
}

type OfflineEvent struct {
	OnlineEvent
	CloseOn int64 `json:"closeon"`
}

type Store interface {
	Online(*OnlineEvent) error
	Offline(*OfflineEvent) error
}
