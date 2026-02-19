module github.com/hannanrazzaghi/wirefight/go-service

go 1.25.0

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/hannanrazzaghi/wirefight/api/proto v0.0.0
	google.golang.org/grpc v1.79.1
)

require (
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/hannanrazzaghi/wirefight/api/proto => ../api/proto
