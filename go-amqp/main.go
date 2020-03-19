package main

import (
	"github.com/mengjayxc/go-amqp/init"
	"github.com/mengjayxc/go-amqp/service/rmq"
)

func main() {
	init.Initialize()

	// TODO 假设两个接收者,通过方法自定义接收者信息
	var aReceiver rmq.Receiver
	var bReceiver rmq.Receiver

	mq := rmq.New()

	// 注册接收者
	mq.RegisterReceiver(aReceiver)
	mq.RegisterReceiver(bReceiver)

	mq.Start()
}
