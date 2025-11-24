package biz

import (
	"context"
	"time"
)

type GiftType int64

const (
	GiftTypeUnknown GiftType = iota
	GiftTypeNormal
	GiftTypeSpecial
)

type Gift struct {
	GiftID     int64     `json:"gift_id"`     // 礼物ID
	SenderID   int64     `json:"sender_id"`   // 发送者ID
	ReceiverID int64     `json:"receiver_id"` // 接收者ID
	Price      int64     `json:"price"`       // 礼物价格
	GiftType   GiftType  `json:"gift_type"`   // 礼物类型
	Quantity   int64     `json:"quantity"`    // 数量
	SendTime   time.Time `json:"send_time"`   // 发送时间
}

type GiftRepo interface {
	Save(ctx context.Context, gift *Gift) (*Gift, error)
	QueryBySender(ctx context.Context, id int64) ([]int64, error)
	QueryByTime(ctx context.Context, startTime time.Time, endTime time.Time) ([]int64, error)
	QueryByValue(ctx context.Context, id int64) ([]int64, error)
	GetGift(ctx context.Context, id int64) (*Gift, error)
	GetTopSenders(ctx context.Context) ([]int64, error)
	GetSendersInLastWeek(ctx context.Context) ([]int64, error)
}

type GiftUsecase struct {
	repo GiftRepo
}

func NewGiftUsecase(repo GiftRepo) GiftUsecase {
	return GiftUsecase{repo: repo}
}
func (uc *GiftUsecase) SendGift(ctx context.Context, senderId int64, receiverId int64, price int32, giftType GiftType, quantity int32) (_r Gift, _err error) {
	return
}

func (uc *GiftUsecase) GetTop10Senders(ctx context.Context) (_r []int64, _err error) {
	return
}
func (uc *GiftUsecase) GetSendersInLastWeek(ctx context.Context) (_r []int64, _err error) {
	return
}
func (uc *GiftUsecase) GetGiftsBySender(ctx context.Context, senderId int64) (_r []*Gift, _err error) {
	return
}
