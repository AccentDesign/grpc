module github.com/accentdesign/grpc/services/email

go 1.21

require (
	github.com/accentdesign/grpc/core v0.0.0
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/google/uuid v1.3.0
	github.com/mocktools/go-smtp-mock/v2 v2.1.0
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.1 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230815205213-6bfd019c3878 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/accentdesign/grpc/core => ./../../core
