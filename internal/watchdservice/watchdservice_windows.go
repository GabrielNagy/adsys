package watchdservice

import (
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

func (s *WatchdService) getServiceArgs() string {
	m, err := lowPrivMgr()
	if err != nil {
		return ""
	}
	defer m.Disconnect()

	svc, err := lowPrivSvc(m, s.Name())
	if err != nil {
		return ""
	}
	defer svc.Close()

	config, err := svc.Config()
	if err != nil {
		return ""
	}

	return config.BinaryPathName
}

func lowPrivMgr() (*mgr.Mgr, error) {
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT|windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return nil, err
	}
	return &mgr.Mgr{Handle: h}, nil
}

func lowPrivSvc(m *mgr.Mgr, name string) (*mgr.Service, error) {
	h, err := windows.OpenService(
		m.Handle, syscall.StringToUTF16Ptr(name),
		windows.SERVICE_QUERY_CONFIG|windows.SERVICE_QUERY_STATUS|windows.SERVICE_START|windows.SERVICE_STOP)
	if err != nil {
		return nil, err
	}
	return &mgr.Service{Handle: h, Name: name}, nil
}
