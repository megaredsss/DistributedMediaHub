package authrepository

type AuthRepository interface {
	// Define methods for the AuthRepository here
	findUserRefreshTokenByID(userID string) (string, error)
}
