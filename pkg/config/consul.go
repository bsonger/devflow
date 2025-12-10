package config

import (
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/model"
	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
	"log"
	"reflect"
	"time"
)

var ConsulClient *api.Client
var lastIndex uint64

func InitConsulClient(consul *model.Consul) error {
	cfg := api.DefaultConfig()
	cfg.Address = consul.Address
	c, err := api.NewClient(cfg)
	if err != nil {
		return err
	}
	ConsulClient = c
	return nil
}

func LoadConsulConfigAndMerge(c *model.Consul) error {
	if ConsulClient == nil {
		return nil
	}

	kv := ConsulClient.KV()
	pair, _, err := kv.Get(c.Key, nil)
	if err != nil {
		return err
	}
	if pair == nil {
		return nil
	}

	cfg := &model.Config{}
	if err := yaml.Unmarshal(pair.Value, cfg); err != nil {
		return err
	}
	MergeConfig(model.C, cfg)
	return nil
}

func MergeConfig(oldCfg, newCfg *model.Config) {
	mergeStructs(oldCfg, newCfg)
}

func mergeStructs(dst, src interface{}) {
	dv := reflect.ValueOf(dst).Elem()
	sv := reflect.ValueOf(src).Elem()

	for i := 0; i < dv.NumField(); i++ {
		df := dv.Field(i)
		sf := sv.Field(i)
		if isNil(sf) {
			continue
		}
		switch df.Kind() {
		case reflect.Struct:
			mergeStructs(df.Addr().Interface(), sf.Addr().Interface())
		default:
			// 覆盖逻辑：src 有值则覆盖
			if !isZero(sf) {
				df.Set(sf)
			}
		}
	}
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Chan, reflect.Func:
		return v.IsNil()
	}
	return false
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return v.Uint() == 0
	}

	return false
}

func WatchConsul(c *model.Consul) {
	if ConsulClient == nil {
		return
	}

	go func() {
		kv := ConsulClient.KV()

		for {
			pair, meta, err := kv.Get(c.Key, &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  60 * time.Second, // 长轮询
			})
			if err != nil {
				log.Println("Consul watch error:", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// 没有变化
			if pair == nil || meta.LastIndex == lastIndex {
				continue
			}

			lastIndex = meta.LastIndex

			logging.Logger.Info("🟢 侦测到 Consul 配置变化，重新加载")

			newCfg := &model.Config{}
			if err := yaml.Unmarshal(pair.Value, newCfg); err != nil {
				log.Println("Consul 配置解析失败:", err)
				continue
			}

			// 覆盖全局配置
			MergeConfig(model.C, newCfg)
			reloadConfig()

		}
	}()
}
