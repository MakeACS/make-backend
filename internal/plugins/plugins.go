package plugins

// generate notifier plugin defs
//go:generate protoc --proto_path=./ --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative -I=./ ./common/plugin.proto
//go:generate protoc --proto_path=./ --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative -I=./ ./notify/interface.proto
//go:generate protoc --proto_path=./ --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative -I=./ ./auth/interface.proto

// generate auth plugin defs

type PluginHTTPForwarding struct {
	Path   string
	ToPort uint16
}
