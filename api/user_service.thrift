namespace go user_service

// 定义用户角色枚举
enum UserRole {
  ADMIN = 1,
  USER = 2,
  GUEST = 3,
}

// 定义地址信息结构
struct Address {
  1: string street,
  2: string city,
  3: string state,
  4: string zipCode,
}

// 定义用户详细信息结构
struct UserDetails {
  1: optional string phone,
  2: optional string website,
  3: optional Address address,
  4: optional map<string, string> metadata,
}

// 用户结构
struct User {
  1: i64 id,
  2: string name,
  3: string email,
  4: i32 age,
  5: UserRole role = UserRole.USER,                    // 枚举类型
  6: optional UserDetails details,                     // 嵌套结构
  7: list<string> tags = [],                           // 列表类型
  8: map<string, string> attributes = {},              // 映射类型
  9: optional bool isActive,                           // 布尔类型
  10: optional double balance,                         // 双精度浮点数
  11: optional i64 createdAt,                          // 长整型
  12: binary avatar,                                   // 二进制数据
}

struct EchoResponse {
  1: i64 serverId,
  2: binary clientData,
  3: User user,
}

service UserService {
  EchoResponse echoData(1: binary clientData, 2: User user),
}