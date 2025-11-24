package service

import (
	"aboveThriftRPC/api/gen-go/user_service"
	"aboveThriftRPC/internal/biz"
	"context"

	"github.com/jinzhu/copier"
)

// UserService is a user service.
type UserService struct {
	uc *biz.UserUsecase
}

// NewUserService new a user service.
func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{
		uc: uc,
	}
}

// 显示构造
func NewThriftUserService(uc *biz.UserUsecase) user_service.UserService {
	return &UserService{
		uc: uc,
	}
}

func (s *UserService) EchoData(ctx context.Context, clientData []byte, user *user_service.User) (_r *user_service.EchoResponse, _err error) {
	echoData, _err := s.uc.Echo(ctx, clientData)
	if _err != nil {
		return
	}
	_r = &user_service.EchoResponse{}
	if err := copier.Copy(_r, echoData); err != nil {
		return nil, err
	}
	return
}
