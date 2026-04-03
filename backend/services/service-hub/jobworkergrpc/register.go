package jobworkergrpc

import (
	jobworkerpb "github.com/gloopai/platform/common/pb/jobworker"
	"github.com/gloopai/platform/service-hub/internal/jobworker"
	"github.com/gloopai/platform/service-hub/internal/store"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Register attaches JobWorkerRuntime gRPC to grpcServer (same process/conn as pay core + ServiceHub).
func Register(grpcServer *grpc.Server, gdb *gorm.DB) {
	st := store.NewScheduledJobsStore(gdb)
	jobworkerpb.RegisterJobWorkerRuntimeServer(grpcServer, jobworker.NewServer(st))
}
