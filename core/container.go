package core

import (
	"fmt"
	"sync"
)

type Container interface {
	Bind(ServiceProvider) error
	IsBind(string) bool

	Make(string) (any, error)
	MustMake(string) any
	MakeNew(string, ...any) (any, error)
}

// 强制要求实现接口
var _ Container = (*EventContainer)(nil)

type EventContainer struct {
	Container
	providers map[string]ServiceProvider
	instances map[string]any
	// 读的频率远高于写
	// Bind是一次性的而Make是频繁的,所以选择RWMutex
	lock sync.RWMutex
}

// NewEventContainer 创建一个服务容器
func NewEventContainer() *EventContainer {
	return &EventContainer{
		providers: map[string]ServiceProvider{},
		instances: map[string]interface{}{},
		lock:      sync.RWMutex{},
	}
}

func (c *EventContainer) Bind(provider ServiceProvider) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := provider.Name()
	c.providers[key] = provider

	if provider.IsDefer() {
		return nil
	}

	if err := provider.Boot(c); err != nil {
		return err
	}

	params := provider.Params(c)
	method := provider.Register(c)
	instance, err := method(params...)
	if err != nil {
		return err
	}
	c.instances[key] = instance
	return nil
}

func (c *EventContainer) IsBind(key string) bool {
	return c.findServiceProvider(key) != nil
}

func (c *EventContainer) findServiceProvider(key string) ServiceProvider {
	c.lock.RLock()
	defer c.lock.RLocker()
	if sp, ok := c.providers[key]; ok {
		return sp
	}
	return nil
}

func (c *EventContainer) Make(key string) (any, error) {
	return c.make(key, nil, false)
}

func (c *EventContainer) MustMake(key string) any {
	provider, err := c.make(key, nil, false)
	if err != nil {
		panic(err)
	}
	return provider
}

func (c *EventContainer) MakeNew(key string, params ...any) (any, error) {
	return c.make(key, params, true)
}

func (c *EventContainer) make(key string, params []any, forceNew bool) (any, error) {
	sp := c.findServiceProvider(key)
	if sp == nil {
		return nil, fmt.Errorf("Container %s have not register", key)
	}

	if forceNew {
		return c.newInstance(sp, params)
	}

	c.lock.RLock()
	defer c.lock.RLocker()

	if sp, ok := c.instances[key]; ok {
		return sp, nil
	}

	ins, err := c.newInstance(sp, params)
	if err != nil {
		return nil, err
	}
	c.instances[key] = ins

	return ins, nil
}

func (c *EventContainer) newInstance(provider ServiceProvider, params []any) (any, error) {
	fmt.Printf("provider => %+v \t %T\n", provider, provider)
	if err := provider.Boot(c); err != nil {
		return nil, err
	}
	if params == nil {
		params = provider.Params(c)
	}

	method := provider.Register(c)

	// a, v := method(params...)
	// fmt.Printf("%+v \t\t\t %+v %+v ????? ", provider, a, v)

	return method(params...)
}
