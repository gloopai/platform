package config

type Config struct {
	Timezone string `json:",optional"`
	Mysql    struct {
		DataSource             string
		MaxOpenConns           int   `json:",optional"`
		MaxIdleConns           int   `json:",optional"`
		ConnMaxLifetimeSeconds int64 `json:",optional"`
	}
	Worker struct {
		ID                   string `json:",optional"`
		PollIntervalSeconds  int64  `json:",optional"`
		MaxClaimPerTick      int64  `json:",optional"`
		MaxEnqueueDuePerTick int64  `json:",optional"`
		// StaleRunningSeconds 超过此时长仍为 running 的行标记为 failed（0=关闭）。防止 worker 崩溃后 forbid 永久阻塞。
		StaleRunningSeconds int64 `json:",optional"`
		// HeartbeatIntervalSeconds 写入 job_worker_heartbeats，供管理台展示节点；0 则默认 15。
		HeartbeatIntervalSeconds int64 `json:",optional"`
	}
}
