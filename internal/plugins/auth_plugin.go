package plugins

type AuthProvider interface {
	Info() PluginInfoResponse
	AuthSomeone(userId int)
}
