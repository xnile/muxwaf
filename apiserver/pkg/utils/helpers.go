package utils

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	tnet "github.com/toolkits/net"
	"io"
	"math"
	"sync"
)

var (
	once     sync.Once
	clientIP = "127.0.0.1"
)

// GenUUID 生成随机字符串，eg: 76d27e8c-a80e-48c8-ad20-e5562e0f67e4
func GenUUID() string {
	u, _ := uuid.NewRandom()
	return u.String()
}

// GetLocalIP 获取本地内网IP
func GetLocalIP() string {
	once.Do(func() {
		ips, _ := tnet.IntranetIP()
		if len(ips) > 0 {
			clientIP = ips[0]
		} else {
			clientIP = "127.0.0.1"
		}
	})
	return clientIP
}

func CheckPageSizeNum(pageNum, pageSize int64) (int64, int64) {
	if pageSize < 1 {
		pageSize = 10
	}
	if pageNum < 1 {
		pageNum = 1
	}
	return pageNum, pageSize
}

func CalPage(count, pageSize int64) int64 {
	return int64(math.Ceil(float64(count) / float64(pageSize)))
}

// MD5 hashes using md5 algorithm
func MD5(text string) string {
	algorithm := md5.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

func Close(c io.Closer) {
	if c == nil {
		return
	}
	if err := c.Close(); err != nil {
		//log.WithError(err).Error("关闭资源文件失败。")
	}
}
