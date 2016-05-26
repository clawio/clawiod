package authenticationcontroller

// AuthenticationController defines an interface to
// grant users access to other services.
type AuthenticationController interface {
	Authenticate(username, password string) (string, error)
}
