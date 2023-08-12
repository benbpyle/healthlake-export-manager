package clients

import (
	"testing"
)

func TestOsFS(t *testing.T) {
	o := OsFS{}
	filename := "/tmp/os_client_test.txt"
	o.Create(filename)
	o.Open(filename)
	o.Remove(filename)
}
