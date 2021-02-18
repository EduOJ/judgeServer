package base

import (
	"context"
	"github.com/go-resty/resty/v2"
	"sync"
)

var HttpClient *resty.Client

var BaseContext, Close = context.WithCancel(context.Background())
var QuitWG = sync.WaitGroup{}
