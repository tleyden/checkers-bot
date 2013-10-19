package checkersbot

type User struct {
	Id         string   `json:"_id"`
	Rev        string   `json:"_rev"`
	TeamId     TeamType `json:"team"`
	GameNumber int      `json:"game"`
}
