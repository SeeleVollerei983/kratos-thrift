package data

import (
	"aboveThriftRPC/internal/biz"
)

type UserRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &UserRepo{
		data: data,
	}
}
