/*
 * Copyright (c) 2026 qecko-labs
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package builder

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/forgezero-cli/ForgeZero/internal/hashpool"
)

var sourceHashes = make(map[string]hashCacheEntry)

func refreshSourceHashes(dirs []string) error {
	return refreshSourceHashesWithCache(dirs, nil)
}

func refreshSourceHashesWithCache(dirs []string, cache map[string]hashCacheEntry) error {
	paths := make([]string, 0)
	metadata := make([]hashCacheEntry, 0)
	for _, root := range dirs {
		if root == "" {
			continue
		}
		if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				fi, serr := os.Stat(path)
				if serr != nil {
					return serr
				}
				if fi.IsDir() {
					return nil
				}
				metadata = append(metadata, hashCacheEntry{size: fi.Size(), modTime: fi.ModTime().UnixNano()})
			} else {
				fi, serr := d.Info()
				if serr != nil {
					return serr
				}
				metadata = append(metadata, hashCacheEntry{size: fi.Size(), modTime: fi.ModTime().UnixNano()})
			}
			paths = append(paths, path)
			return nil
		}); err != nil {
			return err
		}
	}
	result := make(map[string]hashCacheEntry, len(paths))
	if len(paths) == 0 {
		sourceHashes = result
		return nil
	}
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	if workers > len(paths) {
		workers = len(paths)
	}
	type hashJob struct {
		path  string
		index int
	}
	hashes := make([]hashCacheEntry, len(paths))
	pending := 0
	for index, path := range paths {
		if entry, ok := cache[path]; ok && entry.modTime == metadata[index].modTime && entry.size == metadata[index].size && entry.modTime != 0 {
			hashes[index] = entry
			continue
		}
		pending++
	}
	if pending == 0 {
		for index, path := range paths {
			result[path] = hashes[index]
		}
		sourceHashes = result
		return nil
	}
	jobs := make(chan hashJob, workers)
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex
	var stopped atomic.Bool
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 32*1024)
			for job := range jobs {
				if stopped.Load() {
					continue
				}
				f, err := os.Open(job.path)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
					stopped.Store(true)
					continue
				}
				h := hashpool.GetHasher()
				var readErr error
				for {
					n, currentErr := f.Read(buf)
					if n > 0 {
						_, _ = h.Write(buf[:n])
					}
					if currentErr == io.EOF {
						break
					}
					if currentErr != nil {
						readErr = currentErr
						errMu.Lock()
						if firstErr == nil {
							firstErr = currentErr
						}
						errMu.Unlock()
						stopped.Store(true)
						break
					}
				}
				_ = f.Close()
				var sum [32]byte
				h.Sum(sum[:0])
				hashpool.PutHasher(h)
				if readErr == nil {
					hashes[job.index] = hashCacheEntry{
						hash:    sum,
						size:    metadata[job.index].size,
						modTime: metadata[job.index].modTime,
					}
				}
			}
		}()
	}
	for index, path := range paths {
		if hashes[index].modTime == 0 {
			jobs <- hashJob{path: path, index: index}
		}
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}
	for index, path := range paths {
		result[path] = hashes[index]
	}
	sourceHashes = result
	return nil
}
