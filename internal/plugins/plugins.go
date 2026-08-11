package plugins

// generate auth plugin defs
//go:generate protoc --go_out=auth --go-grpc_out=auth --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative -I=./auth ./auth/interface.proto
