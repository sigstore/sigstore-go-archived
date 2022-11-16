//
// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tuf

// CacheKind is used to designate an on-disk or in-memory cache.
type CacheKind int

const (
	Memory CacheKind = iota
	Disk
)

type TUFClientOptions struct {
	// This indicates whether the cache should be in the local filesystem or in-memory.
	// Default: Memory.
	CacheType CacheKind

	// CacheLocation is the location for the local cache.
	// Only applies when CacheType is Disk.
	// This directory will contain the metadata and targets cache for the TUF
	// client.
	CacheLocation string
}

// RepositoryOptions specify options for initializing a particular
// repository in the TUF client.
// Specifies a root.json, a remote, and a name.
//
// TODO: Replace with a map.json for a multi-repository setup.
type RepositoryOptions struct {
	// The trusted root.json
	Root []byte

	// The location of the remote repository.
	Remote string

	// The name of the repository, used to populate the map.json. TODO: Make this
	// optional and use digest of the root.
	Name string
}
