package hubgrpc

import (
	"google.golang.org/grpc"
	"gorm.io/gorm"

	servicehubpb "github.com/gloopai/platform/common/pb/servicehub"
	"github.com/gloopai/platform/service-hub/internal/server"
	"github.com/gloopai/platform/service-hub/internal/svc"
)

// Runtime carries optional NSQ settings for PublishPortalNotification; leave NsqdTCPAddr empty to disable.
// ConsulAddr is used by GetOpsServicesStatus (embedded ServiceHub in pay core must set this to the cluster Consul).
type Runtime struct {
	NsqdTCPAddr       string
	PortalNotifyTopic string
	ConsulAddr        string
}

// Register attaches ServiceHub gRPC to grpcServer using the caller's DB handle (caller owns DSN and pool).
func Register(grpcServer *grpc.Server, gdb *gorm.DB, rt Runtime) {
	ctx := svc.NewServiceContextWithRuntime(gdb, svc.EmbedRuntime{
		NsqdTCPAddr:       rt.NsqdTCPAddr,
		PortalNotifyTopic: rt.PortalNotifyTopic,
		ConsulAddr:        rt.ConsulAddr,
	})
	servicehubpb.RegisterServiceHubServer(grpcServer, server.NewServiceHubServer(ctx))
}
