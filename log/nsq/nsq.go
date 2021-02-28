package nsq

import (
	"errors"
	"io"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/db/mq/mynsq"
	"github.com/cuigh/auxo/util/cast"
)

type writer struct {
	topic string
}

func New(options data.Map) (io.Writer, error) {
	topic := cast.ToString(options.Get("topic"))
	if topic == "" {
		return nil, errors.New("missing required option: topic name")
	}
	w := &writer{topic: topic}
	return w, nil
}

func (w *writer) Write(p []byte) (n int, err error) {

	err = mynsq.MustGetProducer().PublishRaw(w.topic, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
