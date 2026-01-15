module ssh-honeypot

go 1.25

require (
	github.com/NoHaxxJustLags/gimpel/sdk/go v0.0.0
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.44.0
)

require (
	gimpel v0.0.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/NoHaxxJustLags/gimpel/sdk/go => ../../sdk/go

replace gimpel => ../..
