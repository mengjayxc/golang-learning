package rmq

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

// RabbitMQ 用于管理和维护rabbitmq的对象
type RabbitMQ struct {
	wg sync.WaitGroup

	channel      *amqp.Channel
	exchangeName string // exchange名称
	exchangeType string // exchange类型
	receivers    []Receiver
}

//
func New() *RabbitMQ {
	// 这里根据需要自定义
	return &RabbitMQ{
		wg:           sync.WaitGroup{},
		channel:      nil,
		exchangeName: "",
		exchangeType: "",
		receivers:    nil,
	}

}

func (mq *RabbitMQ) prepareExchange() error {
	err := mq.channel.ExchangeDeclare(
		mq.exchangeName,
		mq.exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("exchange Declare err: %s", err.Error())
	}

	return nil
}

// run 开始获取连接并初始化相关操作
func (mq *RabbitMQ) run() {
	// TODO 有刷新连接，重连操作

	// 获取新的channel对象
	//mq.channel =

	// 初始化Exchange
	err := mq.prepareExchange()
	if err != nil {
		logrus.Errorf("exchange Declare err: %s", err)
	}

	for _, receiver := range mq.receivers {
		mq.wg.Add(1)
		// TODO 每个接收者单独启动一个goroutine用来初始化队列并接收消息
		go mq.listen(receiver)
	}

	mq.wg.Wait()

	logrus.Error("所有处理queue的任务都意外退出了")

	//

}

func (mq *RabbitMQ) Start() {
	for {
		mq.run()

		// TODO 一旦连接断开，要隔一段时间去重连
		time.Sleep(3 * time.Second)
	}
}

//
func (mq *RabbitMQ) RegisterReceiver(receiver Receiver) {
	mq.receivers = append(mq.receivers, receiver)
}

// listen 监听指定路由发来的消息
func (mq *RabbitMQ) listen(receiver Receiver) {
	defer mq.wg.Done()

	// 这里获取每个接受者需要监听的队列和路由
	queueName := receiver.QueueName()
	routerKey := receiver.RouterKey()

	// 声明Queue
	_, err := mq.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		// 当队列初始化失败的时候，需要告诉这个接收者相应的错误
		receiver.OnError(fmt.Errorf("初始化队列 %s 失败： %s", queueName, err.Error()))
	}

	// 将Queue绑定到Exchange上
	err = mq.channel.QueueBind(
		queueName,
		routerKey,
		mq.exchangeName,
		false,
		nil,
	)
	if err != nil {
		receiver.OnError(fmt.Errorf("绑定队列 [%s - %s] 到交换机失败: %s", queueName, routerKey, err.Error()))
	}

	// 获取消费通道
	_ = mq.channel.Qos(1, 0, true) // 确保rabbitmq会一个一个发消息
	msgs, err := mq.channel.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		receiver.OnError(fmt.Errorf("获取队列 %s 的消费通道失败: %s", queueName, err.Error()))
	}

	// 使用callback消费数据
	for msg := range msgs {

		for !receiver.OnReceive(msg.Body) {
			logrus.Warn("receiver 数据处理失败，将要重试")
			// TODO
			time.Sleep(1 * time.Second)
		}

		// multiple必须为true, 确保每次只收到一条消息
		msg.Ack(false)

	}

}
