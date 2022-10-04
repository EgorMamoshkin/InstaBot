package tg_client

type UpdatesResponse struct {
	Ok     bool      `json:"ok"`
	Result []Updates `json:"result"`
}
type Updates struct {
	ID      int    `json:"update_id"`
	Message string `json:"message"`
}

type MediaGroup struct {
	contentType string `json:"type"`
	contentURL  string `json:"media"`
	caption     string `json:"caption"`
}
