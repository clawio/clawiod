package simple

import (
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/link/linkcontroller"
	"github.com/satori/go.uuid"
)

type controller struct {
	conf  *config.Config
	links map[string][]*entities.SharedLink
}

// New returns an implementation of LinkController.
func New(conf *config.Config) (linkcontroller.SharedLinkController, error) {
	c := &controller{}
	c.links = make(map[string][]*entities.SharedLink)
	c.conf = conf
	return c, nil
}

func (c *controller) CreateSharedLink(user *entities.User, oinfo *entities.ObjectInfo) (*entities.SharedLink, error) {
	sl := &entities.SharedLink{}
	sl.Token = uuid.NewV4().String()
	sl.Owner = user
	sl.ObjectInfo = oinfo
	c.links[user.Username] = append(c.links[user.Username], sl)
	return sl, nil
}

func (c *controller) ListSharedLinks(user *entities.User) ([]*entities.SharedLink, error) {
	return c.links[user.Username], nil
}
