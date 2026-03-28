// Command devtools：本地联调小工具（与网关进程无关；网关入口见仓库根目录 gateway.go）。
//
//	go run ./cmd/devtools check-pay-invariants
//	go run ./cmd/devtools simulate-channel -order_no=... -paid_amount=... -channel_id=... -secret=...
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "check-pay-invariants":
		runCheckPayInvariants()
	case "simulate-channel", "simulate-upstream":
		runSimulateChannelNotify(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: go run ./cmd/devtools <subcommand> [args]

Subcommands:
  check-pay-invariants   DB invariant checks (payin orders vs fund_logs). Env: PAY_PLATFORM_MYSQL_DSN or MYSQL_*.
  simulate-channel       POST /v1/callback/notify with MD5 sign (OpenAPI base, default :8090). Alias: simulate-upstream.

Examples:
  go run ./cmd/devtools check-pay-invariants
  go run ./cmd/devtools simulate-channel -order_no=Pxxx -paid_amount=100 -channel_id=1 -secret=channel_secret_alt
`)
}
