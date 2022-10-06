package tgclient

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}
type Update struct {
	ID      int         `json:"update_id"`
	Message *IncMessage `json:"message"`
}

type IncMessage struct {
	From User   `json:"from"`
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type User struct {
	UserName string `json:"username"`
}

type Chat struct {
	ID int `json:"id"`
}

type MediaGroup struct {
	ContentType string      `json:"type"`
	ContentURL  string      `json:"media"`
	Caption     interface{} `json:"caption"`
}
