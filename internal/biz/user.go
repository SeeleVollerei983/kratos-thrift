package biz

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

type User struct {
	Id   int64
	Name string
}
type UserRepo interface {
	Save(ctx context.Context, user *User) error
	Get(ctx context.Context, id int64) (*User, error)
	Delete(ctx context.Context, id int64) error
}

type UserUsecase struct {
	repo      UserRepo
	Snowflake *snowflake.Node
}

func NewUserUsecase(repo UserRepo) *UserUsecase {
	node, _ := snowflake.NewNode(1)
	return &UserUsecase{repo: repo, Snowflake: node}
}

type EchoResponse struct {
	ServerId   int64
	ClientData []byte
}

func (uc *UserUsecase) Echo(ctx context.Context, clientData []byte) (*EchoResponse, error) {
	serverId := uc.Snowflake.Generate().Int64()

	// 直接返回原始的 clientData，不需要再次编码
	logrus.Infof("serverId: %d, clientData length: %d", serverId, len(clientData))
	return &EchoResponse{ServerId: serverId, ClientData: clientData}, nil
}
