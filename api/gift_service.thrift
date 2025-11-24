namespace go gift_service

enum GiftType {
  GIFT_TYPE_UNKNOWN = 0,
  GIFT_TYPE_NORMAL = 1,
  GIFT_TYPE_SPECIAL = 2,
}

struct Gift {
  1: i64 giftId,         // 礼物ID
  2: i64 senderId,       // 送礼者ID
  3: i64 receiverId,     // 收礼者ID
  4: i32 price,           // 礼物价格
  5: GiftType giftType,  // 礼物类型
  6: i32 quantity,        // 件数
  7: i64 sendTime,       // 送礼时间（Unix时间戳，单位秒）
}

service GiftService {
  // 送礼操作：发送礼物
  Gift SendGift(1: i64 senderId, 2: i64 receiverId, 3: i32 price, 4: GiftType giftType, 5: i32 quantity),

  // 查询送礼最多的前10人（按累计送礼金额排序）
  list<i64> GetTop10Senders(),

  // 查询一周内送礼的名单（返回送礼者ID列表，去重）
  list<i64> GetSendersInLastWeek(),

  // 查询指定某人所有送礼记录，返回结构体列表
  list<Gift> GetGiftsBySender(1: i64 senderId),
}