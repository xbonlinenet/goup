package demo

type LoginUserInfo struct {
	Uid int64 `json:"uid"`
}

type LoginResponseData struct {
	LoginToken string         `json:"login_token"`
	User       *LoginUserInfo `json:"user"`
}
