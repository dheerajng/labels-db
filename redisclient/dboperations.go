package redisclient

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// PodDBValue is value in DB where key is IP in string format
type PodDBValue struct {
	PodName   string
	Service   string
	Namespace string
	Labels    map[string]string
}

var (
	// Pool is the global var for redis connection pool
	Pool *redis.Pool
)

func init() {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	server := redisHost + ":" + redisPort
	Pool = newPool(server)
	cleanupHook()
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		Pool.Close()
		os.Exit(0)
	}()
}

func getEnv(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultVal
}

// Ping checks if DB is responding or not
func Ping() error {

	conn := Pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		return fmt.Errorf("cannot 'PING' db: %v", err)
	}
	return nil
}

// GetMultiStruct retrieves multiple entries from DB
func GetMultiStruct(multiKeys []string) ([]PodDBValue, error) {
	conn := Pool.Get()
	defer conn.Close()

	var data []PodDBValue
	var keys []interface{}
	for _, k := range multiKeys {
		keys = append(keys, k)
	}
	val, err := redis.ByteSlices(conn.Do("MGET", keys...))
	if err != nil {
		logrus.Errorf("Could not retrive multiple values. %s", err.Error())
		return nil, err
	}
	for _, v := range val {
		d := PodDBValue{}
		_ = json.Unmarshal([]byte(v), &d)
		data = append(data, d)
	}
	fmt.Println("GetMultiStruct: ", data)
	return data, nil
}

// GetStruct retrieves entry from DB
func GetStruct(key string) (PodDBValue, error) {
	conn := Pool.Get()
	defer conn.Close()

	var data PodDBValue
	val, err := redis.String(conn.Do("GET", key))
	if err == redis.ErrNil {
		fmt.Println("Entry does not exist")
	} else if err != nil {
		return data, fmt.Errorf("error getting key %s: %v", key, err)
	}
	err = json.Unmarshal([]byte(val), &data)
	fmt.Println("GetStruct: ", data)
	return data, err
}

// SetStruct creates an entry in DB
func SetStruct(key string, value PodDBValue) error {
	conn := Pool.Get()
	defer conn.Close()

	data, err := json.Marshal(value)
	if err != nil {
		logrus.Errorf("Could not marshal for %s:%v", key, value)
		return err
	}
	_, err = conn.Do("SET", key, data)
	if err != nil {
		return fmt.Errorf("error setting key %s to %s: %v", key, value, err)
	}
	return err
}

// Exists check if entry is present in DB or not
func Exists(key string) (bool, error) {
	conn := Pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
	}
	return ok, err
}

// Delete removes entry from DB
func Delete(key string) error {
	conn := Pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}

/*
func GetKeys(pattern string) ([]string, error) {

	conn := Pool.Get()
	defer conn.Close()

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}

func Incr(counterKey string) (int, error) {

	conn := Pool.Get()
	defer conn.Close()

	return redis.Int(conn.Do("INCR", counterKey))
}
*/
