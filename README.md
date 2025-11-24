# Thrift RPC 微服务示例

这个项目展示了如何使用 Apache Thrift 构建 RPC 微服务，包含了丰富的数据类型和常见用法。

## 目录结构

```
.
├── api/
│   ├── user_service.thrift    # Thrift 接口定义文件
│   └── gen-go/                # 自动生成的 Go 代码
├── client/
│   └── main.go                # 客户端实现
├── server/
│   └── main.go                # 服务端实现
├── go.mod
├── go.sum
└── README.md                  # 本说明文件
```

## 接口特性

Thrift 接口定义包含了以下常见的数据类型和用法：

1. **基本数据类型**: i32, i64, double, string, bool, binary
2. **容器类型**: list, map
3. **枚举类型**: UserRole
4. **结构体嵌套**: User 包含 UserDetails 和 Address 结构
5. **可选字段**: 使用 optional 关键字
6. **默认值**: 为字段设置默认值
7. **自定义异常**: UserNotFoundException, InvalidUserDataException
8. **服务接口**: 完整的 CRUD 操作

## 快速开始

### 1. 安装 Thrift 编译器

确保你已经安装了 Thrift 编译器，可以通过以下方式安装：

```bash
# macOS
brew install thrift

# Ubuntu/Debian
apt-get install thrift-compiler

# 或从源码编译安装
```

### 2. 生成代码

如果需要重新生成 Thrift 代码：

```bash
thrift --gen go -out ./api/gen-go ./api/user_service.thrift
```

### 3. 安装 Go 依赖

```bash
go mod tidy
```

### 4. 运行服务端

打开一个终端窗口，运行服务端：

```bash
go run server/main.go
```

服务端将在 `localhost:9090` 上监听。

### 5. 运行客户端

打开另一个终端窗口，运行客户端：

```bash
go run client/main.go
```

客户端将连接到服务端并执行一系列操作。

## 接口说明

### 数据结构

#### UserRole 枚举
```
enum UserRole {
  ADMIN = 1,
  USER = 2,
  GUEST = 3,
}
```

#### User 结构体
包含了各种常见的数据类型：
- 基本类型：i32, string, bool, double, i64, binary
- 枚举类型：UserRole
- 嵌套结构体：UserDetails
- 容器类型：list<string>, map<string,string>
- 可选字段：使用 optional 标记

#### 异常定义
- UserNotFoundException：用户未找到异常
- InvalidUserDataException：用户数据无效异常

### 服务接口

UserService 提供了完整的 CRUD 操作：

1. `getUser` - 根据 ID 获取用户
2. `createUser` - 创建新用户
3. `updateUser` - 更新用户信息
4. `deleteUser` - 删除用户
5. `listUsers` - 列出用户

每个接口都支持异常处理。

## 功能演示

客户端演示包括：

1. **基本 CRUD 操作**
   - 创建具有完整信息的用户
   - 获取用户详情
   - 更新用户信息
   - 列出所有用户
   - 删除用户

2. **异常处理**
   - 处理用户未找到异常
   - 处理无效数据异常

3. **复杂数据类型使用**
   - 枚举类型的使用
   - 嵌套结构体的操作
   - 列表和映射类型的使用
   - 可选字段的处理

## 开发说明

### 添加新功能

1. 修改 `api/user_service.thrift` 文件
2. 重新生成代码：`thrift --gen go -out ./api/gen-go ./api/user_service.thrift`
3. 更新服务端和客户端实现
4. 运行 `go mod tidy` 更新依赖

### 最佳实践

1. 使用 `optional` 标记非必需字段
2. 为字段提供合适的默认值
3. 合理使用嵌套结构体
4. 定义明确的异常类型
5. 在服务端实现适当的验证逻辑

## 学习要点

通过这个示例你可以学习到：

1. Thrift IDL 的基本语法和高级特性
2. 如何在 Go 中实现 Thrift 服务端和客户端
3. 如何处理各种数据类型
4. 如何实现异常处理机制
5. 微服务间的基本通信模式