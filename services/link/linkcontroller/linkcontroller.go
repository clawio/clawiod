package linkcontroller

import (
	"github.com/clawio/clawiod/entities"
)

// SharedLinkController is an interface to create public shared links.
// This controller nees an authenticated user to create tokens.
// For public consume of shared links, there is PublicSharedLinkController.
type SharedLinkController interface {
	// Authenticated operations
	CreateSharedLink(user *entities.User, oinfo *entities.ObjectInfo, password string, expires int) (*entities.SharedLink, error)
	ListSharedLinks(user *entities.User) ([]*entities.SharedLink, error)
	FindSharedLink(user *entities.User, pathSpec string) (*entities.SharedLink, error)
	DeleteSharedLink(user *entities.User, token string) error

	// Non-Authenticated operations
	IsProtected(token string) (bool, error)
	Info(token, secret string) (*entities.SharedLink, error)
}
