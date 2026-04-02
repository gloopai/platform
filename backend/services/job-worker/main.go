package main

import (
	"flag"
	"os"
	"strings"

	"github.com/gloopai/pay/common/timex"
	"github.com/gloopai/pay/job-worker/internal/config"
	"github.com/gloopai/pay/job-worker/internal/runner"
	"github.com/gloopai/pay/job-worker/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/job-worker.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	// 多实例部署：每个进程唯一 ID，便于 scheduled_job_runs.worker_id 区分。示例见 dev-up.sh。
	if v := strings.TrimSpace(os.Getenv("JOB_WORKER_ID")); v != "" {
		c.Worker.ID = v
	}
	if err := timex.ApplyProcessTimezone(c.Timezone); err != nil {
		panic(err)
	}
	ctx := svc.NewServiceContext(c)
	r := runner.New(ctx)
	r.Start()
}
