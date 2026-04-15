package dto

type SessionQuery struct {
	UserId uint64
}

type SessionMessageQuery struct {
	SessionKey string
}
