package iredisV2

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"strings"
	"time"
)

const (
	defaultAddr         = "127.0.0.1:6379"
	defaultDB           = 0
	defaultMaxRetries   = 3
	defaultPoolSize     = 10 // 默认连接池大小等于 runtime.GOMAXPROCS(cpu个数) * 10
	defaultMinIdleConns = 1  // 启动时，创建最小空闲连接数
)

type Option func(o *options)

type options struct {
	ctx context.Context

	// 客户端连接地址
	// 内建客户端配置，默认为[]string{"127.0.0.1:6379"}
	addrs []string

	// 数据库号
	// 内建客户端配置，默认为0 范围:0~15
	db int

	// 用户名
	// 内建客户端配置，默认为空
	username string

	// 密码
	// 内建客户端配置，默认为空
	password string

	// 最大重试次数
	// 内建客户端配置，默认为3次
	maxRetries int

	// 客户端
	// 外部客户端配置，存在外部客户端时，优先使用外部客户端，默认为nil
	client redis.UniversalClient

	poolSize int

	minIdleConns int

	// 连接超时时间
	// 内建客户端配置，默认为500ms
	dialTimeout time.Duration
}

func defaultOptions() *options {
	return &options{
		ctx:          context.Background(),
		addrs:        []string{defaultAddr},
		db:           defaultDB,
		maxRetries:   defaultMaxRetries,
		poolSize:     100, // 建议直接给一个合理的默认值，或稍微保守一点的计算方式 建议 poolSize 默认给一个固定值（如 50 或 100），除非明确需要，否则不要和 CPU 核心数做乘法，避免集群规模扩大时压垮 Redis。
		minIdleConns: 5,   // 适当增加初始空闲连接，减少冷启动延迟
		username:     "",
		password:     "",
		dialTimeout:  500 * time.Millisecond,
	}
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// WithAddrs 设置连接地址
func WithAddrs(addrs ...string) Option {
	return func(o *options) { o.addrs = addrs }
}

// WithDB 设置数据库号
func WithDB(db int) Option {
	return func(o *options) { o.db = db }
}

// WithUsername 设置用户名
func WithUsername(username string) Option {
	return func(o *options) { o.username = username }
}

/*
你可能会想：“我的 Redis 部署在内网，外网访问不了，明文没关系吧？”
内网渗透风险：很多攻击是先通过 Web 漏洞进入内网，然后在内网横向移动。如果黑客翻到了你机器上的明文配置或代码，你的 Redis 就会被瞬间“秒杀”。
高带宽 = 快速拖庫：在内网高带宽下，一旦密码泄露，黑客可以在几秒钟内导光你所有的缓存数据，甚至利用 Redis 的 SAVE 命令篡改系统文件进行提权。

现代做法（推荐）：
环境变量：代码从 os.Getenv("REDIS_PWD") 读取。
加密存储：配置文件里存的是加密后的密文（如 AES 加密），程序启动时在内存中解密。
配置中心：使用 Nacos、Consul 或 Apollo。这些系统支持权限管控，只有生产环境的机器才有权限获取真实的密码。

安全性：你可以把 ENV:REDIS_MASTER_PWD 写入 config.yaml 并提交到 Git。真正的密码只存在于生产环境的服务器环境变量中（如 K8s 的 Secret），代码库里永远不会出现真密码。
灵活性：它同时兼容了“直接传明文”和“从环境变量读取”两种模式，方便本地调试（直接传）和生产部署（传 ENV:）。
合规性：这能轻松通过 2026 年大部分互联网公司的安全审计。
延伸建议（2026 进阶版）：
如果你的项目使用了 K8s (Kubernetes)，你还可以支持从文件读取（比如读取 /var/run/secrets/redis/password），逻辑是一样的，只需再加一个 strings.HasPrefix(password, "FILE:") 的判断即可。
*/
// WithPassword 设置密码
func WithPassword(password string) Option {
	if strings.HasPrefix(password, "ENV:") {
		envKey := strings.TrimPrefix(password, "ENV:")
		password = os.Getenv(envKey)
		// 优化点：如果环境变量是空的，打印一条警告日志，方便排查连接失败原因
		if password == "" {
			log.Printf("[redis] Warning: environment variable %s is empty\n", envKey)
		}
	}
	return func(o *options) {
		o.password = password
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(o *options) { o.maxRetries = maxRetries }
}

// WithClient 设置外部客户端
func WithClient(client redis.UniversalClient) Option {
	return func(o *options) { o.client = client }
}

func WithPoolSize(poolSize int) Option {
	return func(o *options) { o.poolSize = poolSize }
}

func WithMinIdleConns(minIdleConns int) Option {
	return func(o *options) { o.minIdleConns = minIdleConns }
}

// WithDialTimeout 连接超时
func WithDialTimeout(t time.Duration) Option {
	return func(o *options) {
		// 然后在 redis.UniversalOptions 赋值
		o.dialTimeout = t
	}
}
