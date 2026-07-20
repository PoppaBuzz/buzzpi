package vchiq

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

type ProcessInfo struct {
	PID     string `json:"pid"`
	PPID    string `json:"ppid"`
	Cmd     string `json:"cmd"`
	CPU     string `json:"cpu"`
	Mem     string `json:"mem"`
	Elapsed string `json:"elapsed"`
}

var (
	ErrProcessNotFound = errors.New("process not found")
	ErrGettingProcess  = errors.New("error getting process")
)

func ListProcesses() ([]ProcessInfo, error) {
	cmd := exec.Command("ps", "-e", "-o", "pid,ppid,comm,%cpu,%mem,etime", "--no-headers")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, ErrGettingProcess
	}

	lines := strings.Split(out.String(), "\n")
	var processes []ProcessInfo
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		processes = append(processes, ProcessInfo{
			PID:     fields[0],
			PPID:    fields[1],
			Cmd:     fields[2],
			CPU:     fields[3],
			Mem:     fields[4],
			Elapsed: fields[5],
		})
	}

	if len(processes) == 0 {
		return nil, ErrProcessNotFound
	}
	return processes, nil
}

func GetProcessByPID(pid string) ([]ProcessInfo, error) {
	cmd := exec.Command("ps", "-o", "pid,ppid,cmd,%cpu,etime,%mem", "--pid", pid, "--no-headers")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, ErrGettingProcess
	}

	lines := strings.Split(out.String(), "\n")
	var processes []ProcessInfo
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		processes = append(processes, ProcessInfo{
			PID:     fields[0],
			PPID:    fields[1],
			Cmd:     fields[2],
			CPU:     fields[3],
			Elapsed: fields[4],
			Mem:     fields[5],
		})
	}
	if len(processes) == 0 {
		return nil, ErrProcessNotFound
	}
	return processes, nil
}