package mynsq

import (
	"errors"
	"os"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/cuigh/auxo/util/lazy"
	gonsq "github.com/nsqio/go-nsq"
)

// max_retry=5
// 模型说明
// 系统中单一程序体只可以订阅一个通道，所有人都是同一个通道
// 但是可以对应不同的topic，topic(m)->channel(1)
// 所以channel定义在global config单一app为一个channel
// consumer 消费者结构体
// global:
//       mq:
//         nsq:
//           nsqd_addr:
//           - "127.0.0.1:4150"
//           max_in_flight: 5
//           concurrent: 3
//           max_attempt: 2
//           channel_name: test.nsq
type myConsumer struct {
	isInit      bool
	Debug       bool
	channelName string
	concurrent  int
	maxInFlight int
	maxAttempt  uint16
	//addr 连接地址
	nsqdAddr []string
	// 各个topic的worker
	topics map[string]*topicInfo
}

// topicInfo topic 信息结构体
type topicInfo struct {
	topic         string
	maxInFlight   int
	concurrentNum int
	config        *gonsq.Config
	handler       gonsq.HandlerFunc
	consumer      *gonsq.Consumer
}

// 失败消息处理函数类型
type FailMessageFunc func(message FailMessage) (err error)

func (f FailMessageFunc) HandleFailMessage(message FailMessage) (err error) {
	err = f(message)
	return
}

// 失败消息处理接口,继承了该接口的接口都会调用该接口
type FailMessageHandler interface {
	HandleFailMessage(message FailMessage) (err error)
}

type FailMessage struct {
	Body      []byte
	Attempt   uint16
	Timestamp int64
	MessageID string
	FailMsg   string
}

var (
	mynsqConsumerValue = lazy.Value{New: consumerCreate}
)

// 不定义close
func MustGetConsumer() *myConsumer {
	v, err := mynsqConsumerValue.Get()
	if err != nil {
		logs.Error("MustGetComsumer | must open comsumer failed")
		os.Exit(-1)
	}
	return v.(*myConsumer)
}

func consumerCreate() (d interface{}, err error) {
	options, err := loadOptions()
	if err != nil {
		logs.Error("consumerCreate | loadOptions| err=%v", err)
		return nil, err
	}
	ret := &myConsumer{
		nsqdAddr: make([]string, 0),
		topics:   make(map[string]*topicInfo),
	}
	err = ret.init(options, true)
	if err != nil {
		logs.Error("consumerCreate | ret.init | err=%v", err)
		return nil, err
	}
	d = interface{}(ret)
	return d, err
}

// Connect 连接
func (t *topicInfo) connect(channelName string, nsqdAddr []string, debug bool) {
	if len(nsqdAddr) == 0 {
		logs.Warn("nsqd地址为空，跳过连接,topic:", t.topic)
		return
	}
	var (
		retryNum     = 0
		sleepSeconds = 0
		err          error
	)
	t.consumer, err = gonsq.NewConsumer(t.topic, channelName, t.config)
	if err != nil {
		logs.Error("新建nsq consumer失败，err:%s,topic:%s,channel:%s", err.Error(), t.topic, channelName)
		return
	}
	t.consumer.ChangeMaxInFlight(t.maxInFlight)
	// t.consumer.AddConcurrentHandlers(gonsq.Handler(t.handler), t.concurrentNum)
	t.consumer.AddHandler(gonsq.Handler(t.handler))
	// 不断进行重连，直到连接成功
	for {
		// 只要连上了就不会退出的, 为空判断由入口保证
		if len(nsqdAddr) == 1 {
			err = t.consumer.ConnectToNSQD(nsqdAddr[0])
		} else {
			err = t.consumer.ConnectToNSQDs(nsqdAddr)
		}
		if err != nil {
			logs.Warn("连接nsqd(addr:%v)失败,err:%s", nsqdAddr, err.Error())
			retryNum++
			sleepSeconds = 5
			if retryNum%6 == 0 {
				sleepSeconds = 30
			}
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
			continue
		}
		if debug {
			t.consumer.SetLogger(logs.GetLogger(), gonsq.LogLevelDebug)
		} else {
			t.consumer.SetLogger(logs.GetLogger(), gonsq.LogLevelWarning)
		}
		logs.Info("连接nsqd(%v)成功", nsqdAddr)
		break
	}
	<-t.consumer.StopChan
	err = nil
	return
}

// AddHandler 添加handler
func (c *myConsumer) AddHandler(topic string, handler gonsq.HandlerFunc, failHandler FailMessageFunc) *myConsumer {
	var (
		t  = &topicInfo{}
		ok bool
	)
	if t, ok = c.topics[topic]; !ok {
		t = &topicInfo{}
		t.concurrentNum = c.concurrent
		t.maxInFlight = c.maxInFlight
		t.config = gonsq.NewConfig()
		t.config.MaxAttempts = c.maxAttempt
	}

	t.topic = topic
	// 自定义 handler
	t.handler = func(nm *gonsq.Message) (err error) {
		err = handler(nm)
		if err != nil && c.topics[topic].config.MaxAttempts > 0 && c.topics[topic].config.MaxAttempts == nm.Attempts && failHandler != nil {
			messageID := make([]byte, 0)
			for _, v := range nm.ID {
				messageID = append(messageID, v)
			}
			failHandler(FailMessage{
				MessageID: string(messageID),
				Body:      nm.Body,
				Timestamp: nm.Timestamp,
				FailMsg:   err.Error(),
			})
			err = nil
		}
		return
	}
	c.topics[topic] = t
	return c
}

// StopAll 停止
func (c *myConsumer) stop() {
	for k := range c.topics {
		c.topics[k].consumer.Stop()
	}
}

// Run 运行
func (c *myConsumer) Run() (err error) {
	defer c.stop()
	if !c.isInit {
		err = errors.New("consumer not init")
		return
	}
	if len(c.nsqdAddr) == 0 {
		err = errors.New("nsqd addr address required")
		return
	}
	for _, topicInfo := range c.topics {
		topicInfo.config.MaxAttempts = c.maxAttempt
		topicInfo.config.MaxInFlight = c.maxInFlight
		go topicInfo.connect(c.channelName, c.nsqdAddr, c.Debug)
	}
	neverBack := make(chan int)
	<-neverBack
	return
}

// Init 初始化
func (c *myConsumer) init(configSection *Options, debug bool) (err error) {
	if len(configSection.NsqdAddr) > 0 {
		c.nsqdAddr = configSection.NsqdAddr
	}
	if configSection.MaxInFlight > 0 {
		c.maxInFlight = configSection.MaxInFlight
	}
	if configSection.Concurrent > 0 {
		c.concurrent = configSection.Concurrent
	}
	if configSection.ChannelName != "" {
		c.channelName = configSection.ChannelName
	}
	if c.channelName == "" {
		err = errors.New("config channelName not found")
		return
	}
	if configSection.MaxAttempt > 0 {
		c.maxAttempt = uint16(configSection.MaxAttempt)
	}

	if c.maxInFlight < 1 {
		c.maxInFlight = 1
	}
	if c.concurrent < 1 {
		c.concurrent = 1
	}

	if c.maxInFlight < c.concurrent {
		err = errors.New("max_in_flight should exceed than concurrent")
		return
	}
	c.isInit = true

	return
}
