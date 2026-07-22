package plugins

type UserDataProvider interface {
	FullNameForUser(userID int) (string, error)
	EmailForUser(userID int) (string, error)
}
