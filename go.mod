module github.com/haierspi/pt-gateway

go 1.19

replace github.com/haierspi/pt-gateway/utils => ./utils

require (
	github.com/json-iterator/go v1.1.12
	github.com/lib/pq v1.10.7
	github.com/pborman/uuid v1.2.1
	github.com/robfig/config v0.0.0-20141207224736-0f78529c8c7e
	github.com/streadway/amqp v1.0.0
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
)
