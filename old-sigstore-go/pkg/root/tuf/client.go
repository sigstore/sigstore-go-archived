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

	"github.com/theupdateframework/go-tuf/client"
)

// This is a SigstoreTufClient. Note that this is not opinionated on
// its usage and does not include a sync.Once for single intialization. Users
// of the library are responsible for considering its usage in their application.
// It is threadsafe.
type SigstoreTufClient struct {
	// TODO: Add concurrency support for load operations.

	// client is the base TUF client.
	// TODO: Replace when go-tuf implements a TAP-4 multi-repository client.
	// https://github.com/theupdateframework/go-tuf/issues/348, then this
	// will be an interface.
	client *client.Client

	// local is the TUF local repository for accessing local trusted metadata.
	// TODO: As an optimization, use an in-memory store always, and sync to a
	// configured cache location during updates.
	local client.LocalStore

	// initialized detects whether a remote repository was configured into the
	// TUF client.
	initialized bool
}

// NewSigstoreTufClient creates a new client given client options.
func NewSigstoreTufClient(opts *ClientOptions) (*SigstoreTufClient, error) {
	local, err := localStoreFromOpts(opts)
	if err != nil {
		return nil, err
	}
	return &SigstoreTufClient{local: local}, nil
}

// Initialize initializes the Sigstore TUF Client given a particular repository.
// This WILL run a network call to the remote. The remote may be configured to a
// local filesystem.
// If you intend to load in TrustedRootStore information from fixed information,
// create a new provider.
func (s *SigstoreTufClient) Initialize(opts *RepositoryOptions) error {
	remote, err := remoteStoreFromOpts(opts)
	if err != nil {
		return fmt.Errorf("remoteStoreFromOpts: %w", err)
	}
	s.client = client.NewClient(s.local, remote)
	if err := s.client.Init(opts.Root); err != nil {
		return fmt.Errorf("initializing Sigstore TUF client: %w", err)
	}
	// Update with the TUF client.
	if _, err := s.client.Update(); err != nil {
		return fmt.Errorf("updating Sigstore TUF client: %w", err)
	}
	s.initialized = true
	return nil
}

// GetTrustedRoot returns a TrustedRootStore that can be ingested by verifiers.
func (s *SigstoreTufClient) GetTrustedRootStore() error {
	if !s.initialized {
		// unexpected
		return errors.New("sigstore TUF client must be initialized before usage")
	}

	// This will check the local trusted metadata and assemble the TrustedRootStore
	return errors.New("unimplemented")
}
