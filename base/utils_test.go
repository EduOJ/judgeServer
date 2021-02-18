package base

import (
	"bufio"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
	"strings"
	"testing"
)

func TestStrippedReader(t *testing.T) {
	t.Parallel()
	a := "12\n3123中文测试  \n\n123123中  A文B 测C试\nD😈t est  \n😂0123"
	r := StrippedReader{
		Inner: bufio.NewReader(strings.NewReader(a)),
	}
	p, err := ioutil.ReadAll(&r)
	assert.NoError(t, err)
	assert.Equal(t, "123123中文测试123123中A文B测C试D😈test😂0123", string(p))
}
