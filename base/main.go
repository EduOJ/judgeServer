package base

import (
	"github.com/go-resty/resty/v2"
)

// HC means http client
var HC *resty.Client

var (
	ScriptPath string
	RunPath    string
)
