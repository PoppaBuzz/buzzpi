//go:build linux

package vchiq

import "syscall"

func getMemoryUsagePercent() (float64, error) {
	var sysInfo syscall.Sysinfo_t
	if err := syscall.Sysinfo(&sysInfo); err != nil {
		return 0.0, err
	}
	used := sysInfo.Totalram - sysInfo.Freeram
	return float64(used) / float64(sysInfo.Totalram), nil
}

func getMemory() (float64, float64, float64, error) {
	var sysInfo syscall.Sysinfo_t
	if err := syscall.Sysinfo(&sysInfo); err != nil {
		return 0, 0, 0, err
	}
	total := float64(sysInfo.Totalram)
	free := float64(sysInfo.Freeram)
	used := total - free
	return total, free, used, nil
}

func calculateDiskUsage(path string, percent bool) (float64, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(path, &fs); err != nil {
		return 0, err
	}
	totalSpace := float64(fs.Blocks) * float64(fs.Bsize)
	freeSpace := float64(fs.Bfree) * float64(fs.Bsize)
	usedSpace := totalSpace - freeSpace
	usedPercent := usedSpace / totalSpace

	if percent {
		return usedPercent, nil
	}
	return usedSpace, nil
}

func getDiskSize(path string) (float64, float64, float64, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(path, &fs); err != nil {
		return 0, 0, 0, err
	}
	totalSpace := float64(fs.Blocks) * float64(fs.Bsize)
	freeSpace := float64(fs.Bfree) * float64(fs.Bsize)
	availSpace := float64(fs.Bavail) * float64(fs.Bsize)
	return totalSpace, freeSpace, availSpace, nil
}
