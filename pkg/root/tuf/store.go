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

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/theupdateframework/go-tuf/client"
	tuf_filejsonstore "github.com/theupdateframework/go-tuf/client/filejsonstore"
)

var (
	errUnknownCacheType     = errors.New("unknown cache type")
	errUnknownCacheLocation = errors.New("unknown cache location")
)

// localStoreFromOpts creates a local store depending on the TUF configuration
// and uses the RepositoryOptions to name the metadata directory.
func localStoreFromOpts(opts *TUFClientOptions) (client.LocalStore, error) {
	switch opts.CacheType {
	case Disk:
		if opts.CacheLocation == "" {
			return nil, errUnknownCacheLocation
		}
		return tuf_filejsonstore.NewFileJSONStore(opts.CacheLocation)
	case Memory:
		return client.MemoryLocalStore(), nil
	}
	return nil, errUnknownCacheType
}

// remoteStoreFromOpts creates the remote store using the RepositoryOptions.
// local files may be specified using the file URI scheme.
func remoteStoreFromOpts(repoOpts *RepositoryOptions) (client.RemoteStore, error) {
	u, err := url.ParseRequestURI(repoOpts.Remote)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %s: %w", repoOpts.Remote, err)
	}
	if u.Scheme != "file" {
		return client.HTTPRemoteStore(repoOpts.Remote, nil, nil)
	}
	// Use local filesystem for remote.
	return client.NewFileRemoteStore(os.DirFS(u.Path), "")
}
