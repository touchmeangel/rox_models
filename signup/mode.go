package signup

type SignupMode string

const (
	SignupModeBootstrap  SignupMode = "bootstrap"
	SignupModeOpen       SignupMode = "open"
	SignupModeInviteOnly SignupMode = "invite_only"
)
