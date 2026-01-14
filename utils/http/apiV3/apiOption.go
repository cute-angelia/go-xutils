package apiV3

// Option 定義函數類型
type Option func(*api)

type CryptoType int32

const (
	CryptoTypeNone CryptoType = 0 // 不加密
	CryptoTypeAES  CryptoType = 1 // AES-GCM 加密
	CryptoTypeXOR  CryptoType = 2 // XOR 加密
)

// WithLog 設置日誌開關
func WithLog(on bool) Option {
	return func(a *api) {
		a.isLogOn = on
	}
}

// WithCryptoType 設置加密類型
func WithCryptoType(cryptoType CryptoType) Option {
	return func(a *api) {
		a.cryptoType = cryptoType
	}
}

// WithCryptoKey 設置加密 Key
func WithCryptoKey(cryptoKey string) Option {
	return func(a *api) {
		a.cryptoKey = cryptoKey
	}
}
