//go:build linux

package device

import (
	"bufio"
	"os"
	"strings"
	"syscall"
)

// mountEntry describes a mounted filesystem.
type mountEntry struct {
	path   string
	fstype string
}

// readDiskStats reads disk usage for mounted filesystems on Linux.
func readDiskStats() []DiskStats {
	mounts := linuxMounts()

	var stats []DiskStats
	for _, m := range mounts {
		used, total, avail := diskUsage(m.path)
		if total == 0 {
			continue
		}
		stats = append(stats, DiskStats{
			Mount:       m.path,
			TotalMB:     total,
			UsedMB:      used,
			AvailableMB: avail,
			Percent:     calcPercent(used, total),
		})
	}
	if stats == nil {
		return []DiskStats{}
	}
	return stats
}

// linuxMounts parses /proc/mounts for real filesystem mount points.
func linuxMounts() []mountEntry {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil
	}

	realFS := map[string]bool{
		"ext2": true, "ext3": true, "ext4": true,
		"xfs": true, "btrfs": true, "zfs": true,
		"vfat": true, "ntfs": true, "fuseblk": true,
	}

	seen := make(map[string]bool)
	var mounts []mountEntry

	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		fstype := fields[2]
		mountPoint := fields[1]

		if !realFS[fstype] {
			continue
		}
		if seen[mountPoint] {
			continue
		}
		seen[mountPoint] = true

		if isImportantMount(mountPoint) {
			mounts = append(mounts, mountEntry{path: mountPoint, fstype: fstype})
		}
	}
	return mounts
}

// isImportantMount returns true for mount points we want to report.
func isImportantMount(mountPoint string) bool {
	if mountPoint == "/" {
		return true
	}
	skipPrefixes := []string{"/snap/", "/var/lib/docker", "/media/", "/mnt/"}
	for _, p := range skipPrefixes {
		if strings.HasPrefix(mountPoint, p) {
			return false
		}
	}
	include := map[string]bool{
		"/boot": true, "/home": true, "/var": true,
	}
	return include[mountPoint]
}

// diskUsage returns disk usage for a mount point in MB using statfs.
func diskUsage(path string) (usedMB, totalMB, availMB int64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0
	}
	bsize := int64(stat.Bsize)
	totalMB = int64(stat.Blocks) * bsize / (1024 * 1024)
	availMB = int64(stat.Bavail) * bsize / (1024 * 1024)
	usedMB = (int64(stat.Blocks) - int64(stat.Bfree)) * bsize / (1024 * 1024)
	return
}
