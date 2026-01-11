package conf

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"path/filepath" // 彻底移除 path 包
	"strings"
)

// LoadConfigFile 集成环境变量自动绑定
func LoadConfigFile(cfgFile string) error {
	viper.SetConfigFile(cfgFile)

	// 2026 实践：初始化时执行一次，提升后续 GetEnv 性能
	viper.AutomaticEnv()
	// 关键：将环境变量中的下划线映射到配置的点路径 (如 DB_HOST -> db.host)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Config file not found: %v", err)
		} else {
			log.Printf("Fatal error reading config: %v", err)
		}
	}
	return err
}

func MustLoadConfigFile(cfgFile string) {
	if err := LoadConfigFile(cfgFile); err != nil {
		panic(fmt.Errorf("MustLoadConfigFile failed: %w", err))
	}
}

// LoadConfigByte 优化版
func LoadConfigByte(data []byte, filetype string) error {
	// 2026 推荐：支持 yaml/toml/json/hcl 等多种格式
	viper.SetConfigType(filetype)
	// 使用 bytes.NewReader 性能更优
	return viper.ReadConfig(bytes.NewReader(data))
}

func MustLoadConfigByte(data []byte, filetype string) {
	if err := LoadConfigByte(data, filetype); err != nil {
		panic(fmt.Errorf("MustLoadConfigByte failed: %w", err))
	}
}

// MergeConfig 合并 Reader 配置
func MergeConfig(byteCfg io.Reader) error {
	return viper.MergeConfig(byteCfg)
}

// MergeConfigWithPath 修正版：正确提取目录和文件名
func MergeConfigWithPath(cfgPath string) error {
	dir := filepath.Dir(cfgPath)
	// 去除后缀，Viper MergeInConfig 需要纯文件名
	filename := strings.TrimSuffix(filepath.Base(cfgPath), filepath.Ext(cfgPath))

	viper.AddConfigPath(dir)
	viper.SetConfigName(filename)

	if err := viper.MergeInConfig(); err != nil {
		return fmt.Errorf("merge config failed: %w", err)
	}
	return nil
}

// MustMergeConfigWithPath 修正版：直接复用逻辑，避免逻辑残留
func MustMergeConfigWithPath(cfgPath string) {
	if err := MergeConfigWithPath(cfgPath); err != nil {
		panic(fmt.Errorf("MustMergeConfigWithPath failed: %w", err))
	}
}

// MergeConfigWithMap 合并 Map 配置
func MergeConfigWithMap(cfg map[string]interface{}) error {
	return viper.MergeConfigMap(cfg)
}

// GetEnv 性能优化版
func GetEnv(key string) interface{} {
	return viper.Get(key)
}
