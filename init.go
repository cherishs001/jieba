package jieba

import (
	"github.com/cherishs001/jieba/deps/cppjieba"
	"github.com/cherishs001/jieba/deps/limonp"
	"github.com/cherishs001/jieba/dict"
)

func init() {
	dict.Init()
	limonp.Init()
	cppjieba.Init()
}
