package app

type Update struct {
	ID      int      `json:"update_id"`
	Message *Message `json:"message"`
}

type Message struct {
	ID    int    `json:"message_id"`
	Text  string `json:"text,omitempty"`
	From  *User  `json:"from,omitempty"`
	Chat  *Chat  `json:"chat,omitempty"`
	Audio *Audio `json:"audio"`
	Date  int    `json:"date"`
}

type Chat struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type Audio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
}
