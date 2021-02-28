package mynsq

import (
	"fmt"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/errors"
)

const PkgName = "auxo.mq.nsq"

type Options struct {
	NsqdAddr       []string
	NsqlookupdAddr []string
	MaxInFlight    int
	Concurrent     int
	MaxAttempt     int
	ChannelName    string
}

type interLog struct {
}

func (il *interLog) Output(calldepth int, s string) error {
	fmt.Println(s)
	return nil
}

func loadOptions() (*Options, error) {
	key := "global.mq.nsq"
	if !config.Exist(key) {
		return nil, errors.Format("can't find nsq config for [%s]", key)
	}

	opts := &Options{}
	err := config.UnmarshalOption(key, opts)
	if err != nil {
		return nil, err
	}
	return opts, nil
}
