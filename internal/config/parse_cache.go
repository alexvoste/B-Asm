/*
 *   Copyright (c) 2026 forgezero-cli
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

package config

import (
	"os"
	"path/filepath"
	"sync"
)

type cacheEntry struct {
	modTime int64
	size    int64
	cfg     *Config
}

var (
	configCacheMu sync.RWMutex
	configCache   map[string]*cacheEntry
)

func cacheKey(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return abs
}

func loadConfigCache(path string, fi os.FileInfo) (*Config, bool) {
	configCacheMu.RLock()
	entry := configCache[cacheKey(path)]
	configCacheMu.RUnlock()
	if entry == nil {
		return nil, false
	}
	if entry.size != fi.Size() || entry.modTime != fi.ModTime().UnixNano() {
		return nil, false
	}
	return cloneConfig(entry.cfg), true
}

func storeConfigCache(path string, fi os.FileInfo, cfg *Config) {
	entry := &cacheEntry{
		modTime: fi.ModTime().UnixNano(),
		size:    fi.Size(),
		cfg:     cloneConfig(cfg),
	}
	configCacheMu.Lock()
	if configCache == nil {
		configCache = make(map[string]*cacheEntry)
	}
	configCache[cacheKey(path)] = entry
	configCacheMu.Unlock()
}

func clearConfigCache() {
	configCacheMu.Lock()
	configCache = nil
	configCacheMu.Unlock()
}

func cloneConfig(in *Config) *Config {
	if in == nil {
		return nil
	}
	out := *in
	out.SourceDirs = cloneStringSlice(in.SourceDirs)
	out.SourceFiles = cloneStringSlice(in.SourceFiles)
	out.Exclude = cloneStringSlice(in.Exclude)
	out.Include = cloneStringSlice(in.Include)
	out.Scripts = cloneStringSlice(in.Scripts)
	out.Libs = cloneStringSlice(in.Libs)
	out.AuditIgnore = cloneStringSlice(in.AuditIgnore)
	out.ToolChecksums = cloneStringMap(in.ToolChecksums)
	out.Variables = cloneStringMap(in.Variables)
	out.Flags.Asm = cloneStringSlice(in.Flags.Asm)
	out.Flags.Cc = cloneStringSlice(in.Flags.Cc)
	out.Flags.Ld = cloneStringSlice(in.Flags.Ld)
	out.Preprocess.Inputs = cloneStringSlice(in.Preprocess.Inputs)
	out.Preprocess.Outputs = cloneStringSlice(in.Preprocess.Outputs)
	out.Preprocess.Defines = cloneStringMap(in.Preprocess.Defines)
	out.DepBuild = cloneDepBuild(in.DepBuild)
	out.AutoBuild.BuildOrder = cloneStringSlice(in.AutoBuild.BuildOrder)
	out.AutoBuild.DefaultEnvironment = cloneStringMap(in.AutoBuild.DefaultEnvironment)
	out.ToolchainSettings.SearchPriority = cloneStringSlice(in.ToolchainSettings.SearchPriority)
	out.ToolchainSettings.EnvAllow = cloneStringSlice(in.ToolchainSettings.EnvAllow)
	out.ToolchainSettings.ToolPaths = cloneStringMap(in.ToolchainSettings.ToolPaths)
	out.Hooks = cloneHooks(in.Hooks)
	out.BuildRules = cloneBuildRules(in.BuildRules)
	out.ISO.CustomArgs = cloneStringSlice(in.ISO.CustomArgs)
	return &out
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneHooks(in Hooks) Hooks {
	out := Hooks{OnFailure: in.OnFailure}
	if len(in.PreBuild) == 0 {
		return out
	}
	out.PreBuild = make([]Hook, len(in.PreBuild))
	copy(out.PreBuild, in.PreBuild)
	return out
}

func cloneBuildRules(in []BuildRule) []BuildRule {
	if len(in) == 0 {
		return nil
	}
	out := make([]BuildRule, len(in))
	copy(out, in)
	for i := range out {
		out[i].Inputs = cloneStringSlice(in[i].Inputs)
		out[i].Outputs = cloneStringSlice(in[i].Outputs)
	}
	return out
}

func cloneDepBuild(in DepBuildConfig) DepBuildConfig {
	out := in
	out.BuildTargets = cloneStringSlice(in.BuildTargets)
	out.Outputs = cloneStringSlice(in.Outputs)
	out.Include = cloneStringSlice(in.Include)
	out.Environment = cloneStringMap(in.Environment)
	out.PreBuild = cloneStringSlice(in.PreBuild)
	out.PostBuild = cloneStringSlice(in.PostBuild)
	out.ExcludeFiles = cloneStringSlice(in.ExcludeFiles)
	out.OnlyFiles = cloneStringSlice(in.OnlyFiles)
	out.Steps = cloneBuildSteps(in.Steps)
	if len(in.StepSets) > 0 {
		out.StepSets = make([]StepSet, len(in.StepSets))
		for i := range in.StepSets {
			out.StepSets[i] = in.StepSets[i]
			out.StepSets[i].With = cloneStringMap(in.StepSets[i].With)
			out.StepSets[i].Inputs = cloneStringSlice(in.StepSets[i].Inputs)
			out.StepSets[i].Outputs = cloneStringSlice(in.StepSets[i].Outputs)
		}
	}
	return out
}

func cloneBuildSteps(in []BuildStep) []BuildStep {
	if len(in) == 0 {
		return nil
	}
	out := make([]BuildStep, len(in))
	copy(out, in)
	for i := range out {
		out[i].With = cloneStringMap(in[i].With)
		out[i].Inputs = cloneStringSlice(in[i].Inputs)
		out[i].Outputs = cloneStringSlice(in[i].Outputs)
	}
	return out
}
