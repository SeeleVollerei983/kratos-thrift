package data

import (
	"aboveThriftRPC/internal/biz"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// userRepo implementation of biz.UserRepo interface
type userRepo struct {
	data *Data
}

// NewUserRepo creates a new user repository
func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{
		data: data,
	}
}

// Save saves a user to Redis
func (r *userRepo) Save(ctx context.Context, user *biz.User) error {
	conn := r.data.redis.Get()
	defer conn.Close()

	key := fmt.Sprintf("user:%d", user.Id)

	userJSON, err := json.Marshal(user)
	if err != nil {
		logrus.Errorf("failed to marshal user: %v", err)
		return err
	}

	_, err = conn.Do("SET", key, userJSON)
	if err != nil {
		logrus.Errorf("failed to save user to Redis: %v", err)
		return err
	}

	logrus.Infof("saved user with id: %d", user.Id)
	return nil
}

// Get gets a user from Redis by id
func (r *userRepo) Get(ctx context.Context, id int64) (*biz.User, error) {
	conn := r.data.redis.Get()
	defer conn.Close()

	key := fmt.Sprintf("user:%d", id)

	userJSON, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		logrus.Errorf("failed to get user from Redis: %v", err)
		return nil, err
	}

	var user biz.User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		logrus.Errorf("failed to unmarshal user: %v", err)
		return nil, err
	}

	logrus.Infof("retrieved user with id: %d", id)
	return &user, nil
}

// Delete deletes a user from Redis by id
func (r *userRepo) Delete(ctx context.Context, id int64) error {
	conn := r.data.redis.Get()
	defer conn.Close()

	key := fmt.Sprintf("user:%d", id)

	_, err := conn.Do("DEL", key)
	if err != nil {
		logrus.Errorf("failed to delete user from Redis: %v", err)
		return err
	}

	logrus.Infof("deleted user with id: %d", id)
	return nil
}
