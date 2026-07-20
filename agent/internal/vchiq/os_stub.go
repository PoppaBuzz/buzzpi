//go:build !linux

package vchiq

func getMemoryUsagePercent() (float64, error) { return 0, nil }
func getMemory() (float64, float64, float64, error) { return 0, 0, 0, nil }
func calculateDiskUsage(path string, percent bool) (float64, error) { return 0, nil }
func getDiskSize(path string) (float64, float64, float64, error) { return 0, 0, 0, nil }
