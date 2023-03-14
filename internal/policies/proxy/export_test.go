package proxy

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

// ProxyApplierMock is a mock for the proxy apply object.
type ProxyApplierMock struct {
	WantApplyError bool

	args []string
}

// Call mocks the proxy apply call.
func (d *ProxyApplierMock) Call(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	var errApply error

	for _, arg := range args {
		if arg, ok := arg.(string); ok {
			d.args = append(d.args, arg)
		}
	}

	if d.WantApplyError {
		errApply = errors.New("proxy apply error")
	}

	return &dbus.Call{Err: errApply}
}

func (d *ProxyApplierMock) Args() []string {
	return d.args
}
