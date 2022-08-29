//go:generate protoc  -I ./../../ --go_out ./pb --go-grpc_out ./pb ./../../api/service.proto
package service
