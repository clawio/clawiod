package simple

import (
	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
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

func (c *controller) CreateSharedLink(user *entities.User, oinfo *entities.ObjectInfo, password string, expires int) (*entities.SharedLink, error) {
	sl := &entities.SharedLink{}
	sl.Expires = expires
	sl.Secret = password
	sl.Token = uuid.NewV4().String()
	sl.Owner = user
	sl.ObjectInfo = oinfo
	c.links[user.Username] = append(c.links[user.Username], sl)
	return sl, nil
}

func (c *controller) ListSharedLinks(user *entities.User) ([]*entities.SharedLink, error) {
	return c.links[user.Username], nil
}

func (c *controller) FindSharedLink(user *entities.User, pathSpec string) (*entities.SharedLink, error) {
	return c.getLinkByPathSpec(pathSpec)
}
func (c *controller) DeleteSharedLink(user *entities.User, token string) error {
	for user, links := range c.links {
		for i, link := range links {
			if link.Token == token {
				c.links[user] = append(c.links[user][:i], c.links[user][i+1:]...)
			}
		}

	}
	return nil
}

func (c *controller) IsProtected(token string) (bool, error) {
	link, err := c.getLinkByToken(token)
	if err != nil {
		return false, err
	}
	return c.isLinkProtected(link), nil
}

func (c *controller) Info(token, secret string) (*entities.SharedLink, error) {
	link, err := c.getLinkByToken(token)
	if err != nil {
		return nil, err
	}

	if !c.isSecretCorrect(link, secret) {
		return nil, codes.NewErr(codes.Forbidden, "secret does not match")
	}

	// REDACT secret
	link.Secret = helpers.RedactString(link.Secret)
	return link, nil
}

func (c *controller) getLinkByToken(token string) (*entities.SharedLink, error) {
	for _, links := range c.links {
		for _, link := range links {
			if link.Token == token {
				return link, nil
			}
		}

	}

	return nil, codes.NewErr(codes.NotFound, "link not found")
}

func (c *controller) getLinkByPathSpec(pathSpec string) (*entities.SharedLink, error) {
	for _, links := range c.links {
		for _, link := range links {
			if link.ObjectInfo.PathSpec == pathSpec {
				return link, nil
			}
		}

	}

	return nil, codes.NewErr(codes.NotFound, "link not found")
}
func (c *controller) isLinkProtected(link *entities.SharedLink) bool {
	return link.Secret != ""
}

func (c *controller) isSecretCorrect(link *entities.SharedLink, secret string) bool {
	return link.Secret == secret
}
