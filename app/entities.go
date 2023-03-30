package app

type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type Chat struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type Audio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
}

type Voice Audio

type Message struct {
	ID    int    `json:"message_id"`
	Text  string `json:"text,omitempty"`
	From  *User  `json:"from,omitempty"`
	Chat  *Chat  `json:"chat,omitempty"`
	Audio *Audio `json:"audio"`
	Voice *Voice `json:"voice"`
	Date  int    `json:"date"`
}

type Update struct {
	ID      int      `json:"update_id"`
	Message *Message `json:"message"`
}

type SendMessage struct {
	ChatID    int    `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}
