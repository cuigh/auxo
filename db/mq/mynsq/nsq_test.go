package mynsq_test

import (
	"testing"

	"github.com/astaxie/beego/logs"
	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/db/mq/mynsq"
	gonsq "github.com/nsqio/go-nsq"
)

func init() {
	config.AddFolder("../../../config/samples")
}

func Test_Pulish(t *testing.T) {
	producer := mynsq.MustGetProducer()
	producer.Publish("test_mynsq", "hello world!")
}

func Test_Consumer(t *testing.T) {
	mynsq.MustGetConsumer().AddHandler("test_mynsq", testHandler(), testFailHandler()).Run()
}

func testHandler() gonsq.HandlerFunc {
	return func(nm *gonsq.Message) error {
		logs.Info(string(nm.Body))
		return nil
	}
}

func testFailHandler() mynsq.FailMessageFunc {
	return func(message mynsq.FailMessage) (err error) {
		logs.Error("error msg trigger,msg:", string(message.Body), ",messageid:", message.MessageID)
		err = nil
		return
	}
}
