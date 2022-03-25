package utils

import (
	"net"
	"os"
	"strconv"
	"time"
)

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GetEnvDurationByDefault(env string, unit time.Duration, defaultVal time.Duration) time.Duration {
	val := os.Getenv(env)
	if val == "" {
		return defaultVal
	} else {
		ret, _ := strconv.Atoi(val)
		return time.Duration(ret) * unit
	}
}

func GetEnvUintByDefault(env string, defaultVal uint64) uint64 {
	val := os.Getenv(env)
	if val == "" {
		return defaultVal
	} else {
		ret, _ := strconv.ParseUint(val, 10, 64)
		return ret
	}
}
