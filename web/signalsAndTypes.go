package web

type LoginSignals struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HostSignals struct {
	Name          string `json:"roomName"`
	Locations     string `json:"locations"`
	MaxPlayers    string `json:"maxPlayers"`
	TimerDuration string `json:"timerDuration"`
}

type GameSignals struct {
	Timer  int `json:"timer"`
	Paused int `json:"paused"`
}

type HostRules struct {
	NameTooLong bool
	NameEmpty   bool
}

type SignupRules struct {
	Has8          bool
	UsernameTaken bool
	Created       bool
	LessThan12    bool
}
