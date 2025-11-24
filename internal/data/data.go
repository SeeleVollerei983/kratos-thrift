package data

import (
	"aboveThriftRPC/internal/conf"

	"github.com/google/wire"
	"github.com/sirupsen/logrus"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo)

// Data .
type Data struct {
	// TODO wrapped database client
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	cleanup := func() {
		logrus.Infof("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
