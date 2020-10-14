# xj_game_server

### 目录结构

```
├── README.md                               // 项目描述文件
├── bin                                     // 二进制文件目录
├── client                                  // 模拟客户端
│   ├── client.go
│   └── msg
│       ├── cmd.pb.go
│       └── cmd.proto
├── conf                                    // 配置文件
│   ├── 101_longhudou.yml                   // 101龙虎斗配置
│   ├── global.yml                          // 全局配置
│   ├── log.xml                             // 日志配置
│   ├── login.yml                           // 登录配置
│   ├── rsa_private_key.pem                 // jwt私钥
│   └── rsa_public_key.pem                  // jwt公钥
├── game                                    //  游戏
│   ├── 101_longhudou                       // 101龙虎斗游戏
│   │   ├── Makefile                        // 编译make
│   │   ├── conf                            // 游戏配置
│   │   │   └── conf.go
│   │   ├── db                              // 游戏数据库和缓存
│   │   │   ├── db.go
│   │   │   └── redis.go
│   │   ├── game                            // 游戏业务
│   │   │   ├── external.go
│   │   │   ├── internal
│   │   │   │   ├── chanrpc.go
│   │   │   │   ├── handler.go
│   │   │   │   └── module.go
│   │   │   ├── robot
│   │   │   │   └── robot.go
│   │   │   ├── store
│   │   │   │   └── store.go
│   │   │   ├── table
│   │   │   │   └── table.go
│   │   │   └── user
│   │   │       └── user.go
│   │   ├── gate                            //游戏 路由
│   │   │   ├── external.go
│   │   │   ├── internal
│   │   │   │   └── module.go
│   │   │   └── router.go
│   │   ├── main.go                         // 入口文件
│   │   └── msg                             // 消息结构
│   │       ├── cmd.pb.go
│   │       ├── cmd.proto
│   │       └── msg.go
│   └── public                              // 游戏公用包
│       ├── grpc                            // GRPC
│       │   ├── grpc.go
│       │   ├── grpc.pb.go
│       │   └── grpc.proto
│       ├── mysql                           // Mysql
│       │   └── mysql.go
│       ├── redis                           // redis
│       │   └── redis.go
│       ├── store   
│       │   └── store.go
│       └── user
│           └── user.go
├── go.mod                                  // 包管理
├── go.sum                                  // 包管理
├── logs                                    // 日志打印目录
│   └── access.log.18.11.2019
├── public                                  // 公共包
│   ├── base
│   │   └── skeleton.go
│   ├── config
│   │   └── config.go
│   ├── jwt
│   │   ├── jwt.go
│   │   └── jwt_test.go
│   ├── log
│   │   └── log.go
│   ├── mysql
│   │   └── mysql.go
│   ├── public.go
│   └── redis
│       └── redis.go
├── server                                  // 服务
│   └── hall                               // 大厅心跳服务
│       ├── Makefile                        // Make文件
│       ├── conf                            // 配置文件
│       │   └── conf.go
│       ├── db                              // 数据操作 mysql和缓存
│       │   ├── db.go
│       │   └── redis.go
│       ├── gate                            // 路由
│       │   ├── external.go
│       │   ├── internal
│       │   │   └── module.go
│       │   └── router.go
│       ├── hall                           // 业务
│       │   ├── external.go
│       │   └── internal
│       │       ├── chanrpc.go
│       │       ├── handler.go
│       │       └── module.go
│       ├── main.go                         //启动文件
│       └── msg
│           ├── cmd.pb.go
│           ├── cmd.proto
│           └── msg.go
│   └── login                               // 登录服务
│       ├── Makefile                        // Make文件
│       ├── conf                            // 配置文件
│       │   └── conf.go
│       ├── db                              // 数据操作 mysql和缓存
│       │   ├── db.go
│       │   └── redis.go
│       ├── gate                            // 路由
│       │   ├── external.go
│       │   ├── internal
│       │   │   └── module.go
│       │   └── router.go
│       ├── login                           // 业务
│       │   ├── external.go
│       │   └── internal
│       │       ├── chanrpc.go
│       │       ├── handler.go
│       │       └── module.go
│       ├── main.go                         //启动文件
│       └── msg
│           ├── cmd.pb.go
│           ├── cmd.proto
│           └── msg.go
└── util
    └── leaf                              // 通信包
        ├── ...
```

### 编译运行

1. 登录服务器

```shell script
# 进入login代码目录
cd server/login/
make
# 切换到最外层bin目录
cd ../../bin
./login_server

```


2. 游戏服务器(下面以龙虎斗为例)

```shell script
# 进入login代码目录
cd game/101_longhudou/
make
# 切换到最外层bin目录
cd ../../bin
./101_lhd_server
```