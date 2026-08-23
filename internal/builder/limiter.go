/*
 *   Copyright (c) 2026 qecko-labs
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package builder

import (
	"math"
	"os"
	"runtime"
	"strconv"
)

var compileWorkerMemMB = func() uint64 {
	if v := os.Getenv("FZ_COMPILE_WORKER_MEM_MB"); v != "" {
		if iv, err := strconv.ParseUint(v, 10, 64); err == nil && iv > 0 {
			return iv
		}
	}
	return 1024
}()

func AdjustJobs(requested int) int {
	if requested <= 0 {
		requested = 1
	}
	available := availableMemMB()
	if available == 0 {
		if requested > runtime.NumCPU() {
			requested = runtime.NumCPU()
		}
		return requested
	}
	maxWorkers := int(available / compileWorkerMemMB)
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if requested > maxWorkers {
		requested = maxWorkers
	}
	if requested > runtime.NumCPU() {
		requested = runtime.NumCPU()
	}
	return requested
}

func availableMemMB() uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	available, free := parseMemInfo(data)
	if available > 0 {
		return available / 1024
	}
	return free / 1024
}

func parseMemInfo(data []byte) (uint64, uint64) {
	var available, free uint64
	for len(data) > 0 {
		lineEnd := 0
		for lineEnd < len(data) && data[lineEnd] != '\n' {
			lineEnd++
		}
		key, value, ok := parseMemInfoLine(data[:lineEnd])
		if ok {
			switch key {
			case memInfoAvailable:
				available = value
			case memInfoFree:
				free = value
			}
		}
		if lineEnd == len(data) {
			break
		}
		data = data[lineEnd+1:]
	}
	return available, free
}

const (
	memInfoAvailable = iota + 1
	memInfoFree
)

func parseMemInfoLine(line []byte) (int, uint64, bool) {
	keyLength, key := 0, 0
	if hasPrefix(line, "MemAvailable:") {
		keyLength, key = len("MemAvailable:"), memInfoAvailable
	} else if hasPrefix(line, "MemFree:") {
		keyLength, key = len("MemFree:"), memInfoFree
	} else {
		return 0, 0, false
	}
	valueStart := keyLength
	for valueStart < len(line) && (line[valueStart] == ' ' || line[valueStart] == '\t') {
		valueStart++
	}
	if valueStart == len(line) || line[valueStart] < '0' || line[valueStart] > '9' {
		return 0, 0, false
	}
	var value uint64
	for valueStart < len(line) && line[valueStart] >= '0' && line[valueStart] <= '9' {
		digit := uint64(line[valueStart] - '0')
		if value > (math.MaxUint64-digit)/10 {
			return 0, 0, false
		}
		value = value*10 + digit
		valueStart++
	}
	return key, value, true
}

func hasPrefix(data []byte, prefix string) bool {
	if len(data) < len(prefix) {
		return false
	}
	for index := range prefix {
		if data[index] != prefix[index] {
			return false
		}
	}
	return true
}
