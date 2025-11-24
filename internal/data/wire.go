//go:build wireinject
// +build wireinject

package data

import (
	"aboveThriftRPC/internal/conf"

	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData)