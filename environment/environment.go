package environment

import (
	"os"
	"reflect"
	"strconv"
)

func GetMaxFileSize() (int64, error) {
	return getInt64FromEnv("GOSHARE_MAX_FILE_SIZE")
}

func GetDiskSpace() (int64, error) {
	return getInt64FromEnv("GOSHARE_DISK_SPACE")
}

func GetHost() string {
	env, _ := os.LookupEnv("GOSHARE_HOST")
	return env
}

func ValueOrDefault[T interface{}](val T, def T) T {
	if reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface()) {
		return def
	}
	return val
}

func getInt64FromEnv(key string) (int64, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return 0, nil
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}
