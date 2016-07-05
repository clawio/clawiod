package linkcontroller

import (
	"github.com/clawio/clawiod/entities"
)

// SharedLinkController is an interface to create public shared links.
// This controller nees an authenticated user to create tokens.
// For public consume of shared links, there is PublicSharedLinkController.
type SharedLinkController interface {
	CreateSharedLink(user *entities.User, oinfo *entities.ObjectInfo) (*entities.SharedLink, error)
	ListSharedLinks(user *entities.User) ([]*entities.SharedLink, error)
}
