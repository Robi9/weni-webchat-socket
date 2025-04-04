package config

import (
	"os"
	"testing"
)

var ttDefaultConfigs = Configuration{
	Port:     "8080",
	LogLevel: "info",
	S3: S3{
		AccessKey:      "required",
		SecretKey:      "required",
		Endpoint:       "required",
		Region:         "required",
		Bucket:         "required",
		DisableSSL:     false,
		ForcePathStyle: false,
	},
	RedisQueue: RedisQueue{
		Tag:                   "wwcs-service",
		URL:                   "redis://localhost:6379/1",
		ConsumerPrefetchLimit: 1000,
		ConsumerPollDuration:  100,
		RetryPrefetchLimit:    1000,
		RetryPollDuration:     60000,
	},
}

var ttEnvConfigs = Configuration{
	Port:     "1234",
	LogLevel: "trace",
	S3: S3{
		AccessKey:      "required",
		SecretKey:      "required",
		Endpoint:       "endpoint",
		Region:         "region",
		Bucket:         "bucket",
		DisableSSL:     false,
		ForcePathStyle: false,
	},
	RedisQueue: RedisQueue{
		Tag:                   "wwcs-service",
		URL:                   "redis://localhost:6379/1",
		ConsumerPrefetchLimit: 1000,
		ConsumerPollDuration:  100,
		RetryPrefetchLimit:    1000,
		RetryPollDuration:     60000,
	},
}

var requiredEnvCases = map[string]string{
	"WWC_S3_ACCESS_KEY": "required",
	"WWC_S3_SECRET_KEY": "required",
	"WWC_S3_ENDPOINT":   "required",
	"WWC_S3_REGION":     "required",
	"WWC_S3_BUCKET":     "required",
}

var envCases = map[string]string{
	"WWC_S3_ACCESS_KEY": "required",
	"WWC_S3_SECRET_KEY": "required",
	"WWC_PORT":          "1234",
	"WWC_LOG_LEVEL":     "trace",
	"WWC_S3_ENDPOINT":   "endpoint",
	"WWC_S3_REGION":     "region",
	"WWC_S3_BUCKET":     "bucket",
}

func TestLoadConfigs(t *testing.T) {
	for k, v := range requiredEnvCases {
		os.Setenv(k, v)
	}
	t.Run("Default configs", func(t *testing.T) {
		assertConfigs(t, ttDefaultConfigs)
	})

	t.Run("Env configs", func(t *testing.T) {
		for k, v := range envCases {
			os.Setenv(k, v)
			defer os.Unsetenv(k)
		}
		assertConfigs(t, ttEnvConfigs)
	})
}

func assertConfigs(t *testing.T, want Configuration) {
	t.Helper()

	have := loadConfigs()
	if have != want {
		t.Errorf("\nhave %#v, \nwant %#v", have, want)
	}
}
