package mynsq

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/util/lazy"
	gonsq "github.com/nsqio/go-nsq"
)

var (
	mynsqProducerValue = lazy.Value{New: create}
)

type myProducer struct {
	producer *gonsq.Producer
}

// 不定义close
func MustGetProducer() *myProducer {
	v, err := mynsqProducerValue.Get()
	if err != nil {
		logs.Error("MustGetProducer | must open producer failed")
		os.Exit(-1)
	}
	return v.(*myProducer)
}

func create() (d interface{}, err error) {
	options, err := loadOptions()
	if err != nil {
		return nil, err
	}
	if len(options.NsqdAddr) <= 0 {
		return nil, errors.New("create | options.NsqdAddr is empty")
	}
	config := gonsq.NewConfig()
	w, err := gonsq.NewProducer(options.NsqdAddr[0], config)
	if err != nil {
		err = errors.New("初始化 nsq producer 失败, err:" + err.Error())
		return
	}
	w.SetLogger(logs.GetLogger(), gonsq.LogLevelDebug)
	ret := &myProducer{producer: w}
	d = interface{}(ret)
	return d, err
}

// marshalMsg 将消息解析成[]byte,如果出错,第二个参数返回 error
func (p *myProducer) marshalMsg(msg interface{}) (m []byte, err error) {
	switch t := msg.(type) {
	case []byte:
		m = t
	case float64:
		m = []byte(strconv.FormatFloat(t, 'f', -1, 64))
	case int64:
		m = []byte(strconv.FormatInt(t, 10))
	case string:
		m = []byte(t)
	default:
		m, err = json.Marshal(msg)
	}

	return
}

// Publish 投递消息,如果失败,返回 error
func (p *myProducer) Publish(topic string, msg interface{}) (err error) {
	var (
		m []byte
	)
	if m, err = p.marshalMsg(msg); err != nil {
		return
	}
	err = p.producer.Publish(topic, m)

	return
}

// Publish 投递消息,如果失败,返回 error
func (p *myProducer) PublishRaw(topic string, m []byte) (err error) {

	err = p.producer.Publish(topic, m)

	return
}

// MultiPublish 批量发布消息,如果失败,返回 error
func (p *myProducer) MultiPublish(topic string, msg [][]interface{}) (err error) {
	var (
		m   = make([][]byte, 0)
		tmp []byte
	)
	for _, v := range msg {
		if tmp, err = p.marshalMsg(v); err != nil {
			return
		}
		m = append(m, tmp)
	}
	err = p.producer.MultiPublish(topic, m)

	return
}

func (p *myProducer) DeferPublish(topic string, msg interface{}, deferSecond int64) (err error) {
	var (
		m []byte
	)
	if m, err = p.marshalMsg(msg); err != nil {
		return
	}
	err = p.producer.DeferredPublish(topic, time.Second*time.Duration(deferSecond), m)
	return
}
