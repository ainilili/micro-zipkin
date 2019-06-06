# Summary
github.com/micro/go-plugins is very huge, and contains a log of package dependency. 
This repository is only a zipkin opentracing wapper for go-micro project, so the usage of this module is more simple and lightweight.

# Usage
```
package main

import (
	"github.com/micro/go-micro"
	"github.com/x-punch/micro-zipkin"
)

func main() {
	zipkin.SetGlobalTracer("test.srv", ":80", "https://zipkin/api/v1/spans")
	service := micro.NewService(micro.WrapHandler(zipkin.NewHandlerWrapper()))
	service.Init(micro.Name("test.srv"), micro.Address(":80"))
	if err := service.Run(); err != nil {
		panic("Failed to run service: " + err.Error())
	}
}
```

