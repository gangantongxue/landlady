package version

import (
	"testing"

	"github.com/gangantongxue/landlady/cmd/test"
)

func TestInit(t *testing.T) {
	test.Setup(t)
	Init()
	t.Log("version: ", GetVersion())
	t.Log("author: ", GetAuthor())
	t.Log("email: ", GetEmail())
	PrintVersion()
}
