package main

import (
	"fmt"
	"ioc/core"
	"ioc/provider/demo"
)

func main() {
	c := core.NewEventContainer()

	c.Bind(&demo.DemoServiceProvider{})

	fmt.Println("Bind 结束")
	// 获取demo服务实例
	demoService := c.MustMake(demo.Key).(demo.Service)

	// 调用服务实例的方法
	foo := demoService.GetFoo()

	fmt.Println(foo)
}
