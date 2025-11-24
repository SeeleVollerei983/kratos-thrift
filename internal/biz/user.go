package biz

import (
	"context"
	"encoding/json"

	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

type User struct {
	Id   int64
	Name string
}
type UserRepo interface{}

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

	jsonData, err := json.Marshal(clientData)
	if err != nil {
		logrus.Errorf("marshal clientData failed, err:%v, clientData:%v", err, clientData)
		return nil, err
	}
	logrus.Infof("serverId: %d, clientData: %s", serverId, jsonData)
	return &EchoResponse{ServerId: serverId, ClientData: jsonData}, nil
}
