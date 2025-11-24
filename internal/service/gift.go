package service

import (
	"aboveThriftRPC/api/gen-go/gift_service"
	"aboveThriftRPC/internal/biz"
	"context"

	"github.com/jinzhu/copier"
)

type GiftService struct {
	Uc biz.GiftUsecase
}

func NewGiftService(uc biz.GiftUsecase) *GiftService {
	return &GiftService{
		Uc: uc,
	}
}

// 显示构造
func NewThriftGiftService(uc biz.GiftUsecase) gift_service.GiftService {
	return &GiftService{
		Uc: uc,
	}
}

func (s *GiftService) SendGift(ctx context.Context, senderId int64, receiverId int64, price int32, giftType gift_service.GiftType, quantity int32) (_r *gift_service.Gift, _err error) {
	gift, err := s.Uc.SendGift(ctx, senderId, receiverId, price, biz.GiftType(giftType), quantity)
	if err != nil {
		return nil, err
	}
	if err := copier.Copy(&_r, gift); err != nil {
		return nil, err
	}
	return _r, nil
}

func (s *GiftService) GetTop10Senders(ctx context.Context) (_r []int64, _err error) {
	return s.Uc.GetTop10Senders(ctx)
}

func (s *GiftService) GetSendersInLastWeek(ctx context.Context) (_r []int64, _err error) {
	_r, err := s.Uc.GetSendersInLastWeek(ctx)
	if err != nil {
		return nil, err
	}
	return _r, nil
}

func (s *GiftService) GetGiftsBySender(ctx context.Context, senderId int64) (_r []*gift_service.Gift, _err error) {
	gifts, err := s.Uc.GetGiftsBySender(ctx, senderId)
	if err != nil {
		return nil, err
	}
	if err := copier.Copy(&_r, gifts); err != nil {
		return nil, err
	}
	return _r, nil
}
