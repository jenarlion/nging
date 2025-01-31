package system

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/webx-top/com"
)

type Process struct {
	Name       string  `json:"name"`
	Pid        int32   `json:"pid"`
	Ppid       int32   `json:"ppid"`
	CPUPercent float64 `json:"cpuPercent"`
	MemPercent float32 `json:"memPercent"`
	MemUsed    uint64  `json:"memUsed"`
	//Running    bool    `json:"running"`
	CreateTime string `json:"createTime"`
	created    int64
	Exe        string   `json:"exe"`
	Cmdline    string   `json:"cmdline"`
	Cwd        string   `json:"cwd"`
	Status     []string `json:"status"`
	Username   string   `json:"username"`
	NumThreads int32    `json:"numThreads"`
	NumFDs     int32    `json:"numFDs"`
}

func (p *Process) Parse(ctx context.Context, proc *process.Process) (*Process, error) {
	var err error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(`%v`, e)
		}
	}()
	p.Pid = proc.Pid
	p.CPUPercent, err = proc.CPUPercentWithContext(ctx)
	if err != nil {
		return p, err
	}
	//p.Running, err = proc.IsRunningWithContext(ctx)
	p.created, err = proc.CreateTimeWithContext(ctx)
	if err != nil {
		return p, err
	}
	if p.created > 0 {
		p.CreateTime = com.DateFormat(`Y-m-d H:i:s`, p.created/1000)
	}
	p.MemPercent, err = p.MemoryPercentWithContext(ctx, proc)
	if err != nil {
		return p, err
	}
	p.Ppid, err = proc.PpidWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.Name, err = proc.NameWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.Exe, err = proc.ExeWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.Cmdline, err = proc.CmdlineWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.Cwd, err = proc.CwdWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.Status, err = proc.StatusWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.Username, err = proc.UsernameWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.NumThreads, err = proc.NumThreadsWithContext(ctx)
	if err != nil {
		return p, err
	}
	p.NumFDs, err = proc.NumFDsWithContext(ctx)
	return p, err
}

func (p *Process) MemoryPercentWithContext(ctx context.Context, proc *process.Process) (float32, error) {
	var err error
	machineMemory, ok := ctx.Value(`system.machineMemory`).(*mem.VirtualMemoryStat)
	if !ok {
		machineMemory, err = mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			return 0, err
		}
		ctx = context.WithValue(ctx, `system.machineMemory`, machineMemory)
	}
	total := machineMemory.Total

	processMemory, err := proc.MemoryInfoWithContext(ctx)
	if err != nil {
		return 0, err
	}
	used := processMemory.RSS

	p.MemUsed = used // set
	if total == 0 {
		return 0, nil
	}
	return (100 * float32(used) / float32(total)), nil
}

func ProcessList(ctx context.Context) ([]*Process, error) {
	var err error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(`%v`, e)
		}
	}()
	var list []*process.Process
	list, err = process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	processes := make([]*Process, len(list))
	exec := func(idx int, proc *process.Process) (err error) {
		p := &Process{}
		processes[idx], err = p.Parse(ctx, proc)
		return
	}
	for idx, proc := range list {
		err = exec(idx, proc)
		if err != nil {
			return processes, err
		}
	}
	return processes, err
}
