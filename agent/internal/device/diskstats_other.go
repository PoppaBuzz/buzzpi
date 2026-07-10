//go:build !linux

package device

// readDiskStats returns an empty disk stats slice on non-Linux platforms.
func readDiskStats() []DiskStats {
	return []DiskStats{}
}
