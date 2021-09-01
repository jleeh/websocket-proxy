module main

go 1.16

require (
	github.com/aws/aws-sdk-go v1.40.34 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jleeh/websocket-proxy v0.0.0-20190402011646-1724798393d8
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.1 // indirect
)

replace github.com/jleeh/websocket-proxy => ../
