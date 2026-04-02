package hubgrpc

import (
	"google.golang.org/grpc"
	"gorm.io/gorm"

	servicehubpb "github.com/gloopai/platform/common/pb/servicehub"
	"github.com/gloopai/platform/service-hub/internal/server"
	"github.com/gloopai/platform/service-hub/internal/svc"
)

// Runtime carries optional NSQ settings for PublishPortalNotification; leave NsqdTCPAddr empty to disable.
type Runtime struct {
	NsqdTCPAddr       string
	PortalNotifyTopic string
}

// Register attaches ServiceHub gRPC to grpcServer using the caller's DB handle (caller owns DSN and pool).
func Register(grpcServer *grpc.Server, gdb *gorm.DB, rt Runtime) {
	ctx := svc.NewServiceContextWithRuntime(gdb, rt.NsqdTCPAddr, rt.PortalNotifyTopic)
	servicehubpb.RegisterServiceHubServer(grpcServer, server.NewServiceHubServer(ctx))
}
