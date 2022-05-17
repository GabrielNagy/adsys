package commands

import (
	"fmt"
	"golang.org/x/sys/windows"
)

const processEntrySize = 568

func pidCount(name string) (int, error) {
	name += ".exe"
	fmt.Println("Searching for", name)
	var count int
	h, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return 0, e
	}
	p := windows.ProcessEntry32{Size: processEntrySize}
	for {
		e := windows.Process32Next(h, &p)
		if e != nil {
			return count, e
		}
		fmt.Println(windows.UTF16ToString(p.ExeFile[:]))
		if windows.UTF16ToString(p.ExeFile[:]) == name {
			count++
		}
	}
	return count, nil
}
