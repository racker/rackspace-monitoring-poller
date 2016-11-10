package main

import (
	//"github.com/shirou/gopsutil/cpu"
	log "github.com/Sirupsen/logrus"
)

type HostInfoCpu struct {
	HostInfoBase
}

type HostInfoCpuResult struct {

}

func NewHostInfoCpu(base *HostInfoBase) HostInfo {
	return &HostInfoCpu{HostInfoBase: *base}
}

func (*HostInfoCpu) Run() (*CheckResult, error) {
	log.Println("Running CPU")
	/*
		stats, err := cpu.Times(false)
		if err != nil {
			return nil, err
		}
		info, err := cpu.Info()
		if err != nil {
			return nil, err
		}
	*/
	cr := NewCheckResult()
	//TODO
	return cr, nil
}

func (*HostInfoCpu) BuildResult(cr *CheckResult) interface{}  {
	//TODO
	return nil
}
