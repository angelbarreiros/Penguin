package router

import (
	"sync"
)

var routerInstance *router
var routerOnce sync.Once
