module github.com/gloopai/pay/notice-consumer

go 1.25.1

require (
	github.com/gloopai/pay/common v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.9.3
	github.com/nsqio/go-nsq v1.1.0
)

replace github.com/gloopai/pay/common => ../../common

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
)
