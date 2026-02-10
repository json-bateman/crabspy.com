package web

type LoginSignals struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HostSignals struct {
	Name       string `json:"roomName"`
	MaxPlayers string `json:"maxPlayers"`
	Locations  string `json:"locations"`
	Private    string `json:"private"`
}
