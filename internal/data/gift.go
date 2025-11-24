package data

import (
	"aboveThriftRPC/internal/biz"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// giftRepo implementation of biz.GiftRepo interface
type GiftRepo struct {
	data *Data
}

// NewGiftRepo creates a new gift repository
func NewGiftRepo(data *Data) biz.GiftRepo {
	return &GiftRepo{
		data: data,
	}
}

// Save saves a gift to Redis
func (r *GiftRepo) Save(ctx context.Context, gift *biz.Gift) (*biz.Gift, error) {
	conn := r.data.redis.Get()
	defer conn.Close()

	// Store gift details
	giftKey := fmt.Sprintf("gift:%d", gift.GiftID)
	giftJSON, err := json.Marshal(gift)
	if err != nil {
		logrus.Errorf("failed to marshal gift: %v", err)
		return nil, err
	}

	_, err = conn.Do("SET", giftKey, giftJSON)
	if err != nil {
		logrus.Errorf("failed to save gift to Redis: %v", err)
		return nil, err
	}

	// Add to sender's gifts list
	senderKey := fmt.Sprintf("sender:%d:gifts", gift.SenderID)
	_, err = conn.Do("SADD", senderKey, gift.GiftID)
	if err != nil {
		logrus.Errorf("failed to add gift to sender's list: %v", err)
		return nil, err
	}

	// Add to time-based sorted set
	timeKey := "gifts:by_time"
	timestamp := float64(gift.SendTime.Unix())
	_, err = conn.Do("ZADD", timeKey, timestamp, gift.GiftID)
	if err != nil {
		logrus.Errorf("failed to add gift to time-based set: %v", err)
		return nil, err
	}

	// Add to value-based sorted set
	valueKey := "gifts:by_value"
	_, err = conn.Do("ZADD", valueKey, gift.Price, gift.GiftID)
	if err != nil {
		logrus.Errorf("failed to add gift to value-based set: %v", err)
		return nil, err
	}

	logrus.Infof("saved gift with id: %d", gift.GiftID)
	return gift, nil
}

// QueryBySender returns gift IDs sent by a specific sender
func (r *GiftRepo) QueryBySender(ctx context.Context, id int64) ([]int64, error) {
	conn := r.data.redis.Get()
	defer conn.Close()

	senderKey := fmt.Sprintf("sender:%d:gifts", id)

	giftIDs, err := redis.Int64s(conn.Do("SMEMBERS", senderKey))
	if err != nil {
		logrus.Errorf("failed to query gifts by sender: %v", err)
		return nil, err
	}

	logrus.Infof("found %d gifts for sender: %d", len(giftIDs), id)
	return giftIDs, nil
}

// QueryByTime returns gift IDs sent within a time range
func (r *GiftRepo) QueryByTime(ctx context.Context, startTime time.Time, endTime time.Time) ([]int64, error) {
	conn := r.data.redis.Get()
	defer conn.Close()

	timeKey := "gifts:by_time"

	startTimestamp := float64(startTime.Unix())
	endTimestamp := float64(endTime.Unix())

	giftIDs, err := redis.Int64s(conn.Do("ZRANGEBYSCORE", timeKey, startTimestamp, endTimestamp))
	if err != nil {
		logrus.Errorf("failed to query gifts by time: %v", err)
		return nil, err
	}

	logrus.Infof("found %d gifts in time range %v to %v", len(giftIDs), startTime, endTime)
	return giftIDs, nil
}

// QueryByValue returns gift IDs with value greater than or equal to the given value
func (r *GiftRepo) QueryByValue(ctx context.Context, id int64) ([]int64, error) {
	conn := r.data.redis.Get()
	defer conn.Close()

	valueKey := "gifts:by_value"

	// Use the id as minimum value threshold
	giftIDs, err := redis.Int64s(conn.Do("ZRANGEBYSCORE", valueKey, id, "+inf"))
	if err != nil {
		logrus.Errorf("failed to query gifts by value: %v", err)
		return nil, err
	}

	logrus.Infof("found %d gifts with value >= %d", len(giftIDs), id)
	return giftIDs, nil
}

// GetGift retrieves a gift by ID
func (r *GiftRepo) GetGift(ctx context.Context, id int64) (*biz.Gift, error) {
	conn := r.data.redis.Get()
	defer conn.Close()
	giftKey := fmt.Sprintf("gift:%d", id)

	giftJSON, err := redis.Bytes(conn.Do("GET", giftKey))
	if err != nil {
		if err == redis.ErrNil {
			return nil, fmt.Errorf("gift with id %d not found", id)
		}
		logrus.Errorf("failed to get gift from Redis: %v", err)
		return nil, err
	}

	var gift biz.Gift
	err = json.Unmarshal(giftJSON, &gift)
	if err != nil {
		logrus.Errorf("failed to unmarshal gift: %v", err)
		return nil, err
	}

	logrus.Infof("retrieved gift with id: %d", id)
	return &gift, nil
}

// GetTopSenders returns top 10 senders by total gift value
func (r *GiftRepo) GetTopSenders(ctx context.Context) ([]int64, error) {
	conn := r.data.redis.Get()
	defer conn.Close()

	// This is a simplified implementation
	// In a real scenario, you might maintain a separate sorted set for sender totals
	// Here we'll use a Lua script to calculate totals on the fly

	script := `
	local senderKeys = redis.call('KEYS', 'sender:*:gifts')
	local senderTotals = {}
	
	for i=1,#senderKeys do
		local giftIds = redis.call('SMEMBERS', senderKeys[i])
		local total = 0
		
		for j=1,#giftIds do
			local giftKey = 'gift:' .. giftIds[j]
			local giftData = redis.call('GET', giftKey)
			if giftData then
				local gift = cjson.decode(giftData)
				total = total + gift.price
			end
		end
		
		local senderId = string.match(senderKeys[i], 'sender:(%d+):gifts')
		table.insert(senderTotals, {tonumber(senderId), total})
	end
	
	-- Sort by total value in descending order
	table.sort(senderTotals, function(a, b) return a[2] > b[2] end)
	
	-- Return top 10 sender IDs
	local result = {}
	for i=1, math.min(10, #senderTotals) do
		table.insert(result, senderTotals[i][1])
	end
	
	return result
	`

	result, err := redis.Int64s(conn.Do("EVAL", script, 0))
	if err != nil {
		logrus.Errorf("failed to get top senders: %v", err)
		return nil, err
	}

	logrus.Infof("found %d top senders", len(result))
	return result, nil
}

// GetSendersInLastWeek returns sender IDs who sent gifts in the last week
func (r *GiftRepo) GetSendersInLastWeek(ctx context.Context) ([]int64, error) {
	// Get gifts from the last week
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)

	giftIDs, err := r.QueryByTime(ctx, weekAgo, now)
	if err != nil {
		return nil, err
	}

	// Get unique sender IDs from these gifts
	senderSet := make(map[int64]bool)
	for _, giftID := range giftIDs {
		gift, err := r.GetGift(ctx, giftID)
		if err != nil {
			logrus.Errorf("failed to get gift %d: %v", giftID, err)
			continue
		}
		senderSet[gift.SenderID] = true
	}

	// Convert set to slice
	var senders []int64
	for senderID := range senderSet {
		senders = append(senders, senderID)
	}

	logrus.Infof("found %d senders in the last week", len(senders))
	return senders, nil
}
