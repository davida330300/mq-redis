package rediskeys

import "time"

const (
	JobKeyPrefix     = "job:"
	JobDataKeyPrefix = "job:data:"

	RetryJobsKey = "retry:jobs"
	RetryLockKey = "retry:lock"
)

const (
	DedupeTTL    = 72 * time.Hour
	JobStatusTTL = 14 * 24 * time.Hour
	JobDataTTL   = 14 * 24 * time.Hour
	DLQTTL       = 14 * 24 * time.Hour
)

func JobKey(id string) string {
	return JobKeyPrefix + id
}

func JobDataKey(id string) string {
	return JobDataKeyPrefix + id
}
