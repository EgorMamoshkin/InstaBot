package instagramapi

type User struct {
	UserID int    `json:"user_id"`
	Token  string `json:"access_token"`
}

type UserMedia struct {
	Data []Media `json:"data"`
}

type Media struct {
	ID        string     `json:"id"`
	Caption   string     `json:"caption"`
	MediaType string     `json:"media_type"`
	MediaURL  string     `json:"media_url"`
	Username  string     `json:"username"`
	Timestamp string     `json:"timestamp"`
	Children  *UserMedia `json:"children"`
}
