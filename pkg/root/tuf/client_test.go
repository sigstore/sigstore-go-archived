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
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// TODO(asraa): Add support for:
//   - expired timestamps and other repository states with embedded testdata.
//   - concurrency
func TestInitialize(t *testing.T) {
	t.Parallel()
	td := t.TempDir()

	// Create a new TUF repository
	testRepo := newTufRepository(t, td)
	testRepo.addTarget("foo.txt", []byte("hello"), nil)
	testRepo.publish()
	rootBytes := testRepo.root()

	// Serve remote repository.
	s := httptest.NewServer(http.FileServer(http.Dir(filepath.Join(td, "repository"))))
	t.Cleanup(func() {
		s.Close()
	})

	testCases := []struct {
		name          string
		tufOpts       *ClientOptions
		repoOpts      *RepositoryOptions
		wantClientErr bool
		wantInitErr   bool
	}{
		{
			name: "fail: unknown cache type",
			tufOpts: &ClientOptions{
				CacheType: 3,
			},
			repoOpts:      &RepositoryOptions{},
			wantClientErr: true,
		},
		{
			name: "valid memory cache with fs remote",
			tufOpts: &ClientOptions{
				CacheType: Memory,
			},
			repoOpts: &RepositoryOptions{
				Name:   "sigstore-staging",
				Remote: fmt.Sprintf("file://%s/repository", td),
				Root:   rootBytes,
			},
		},
		{
			name: "valid memory cache with http remote",
			tufOpts: &ClientOptions{
				CacheType: Memory,
			},
			repoOpts: &RepositoryOptions{
				Name:   "sigstore-staging",
				Remote: s.URL,
				Root:   rootBytes,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, clientErr := NewSigstoreTufClient(tc.tufOpts)
			if clientErr != nil {
				if !tc.wantClientErr {
					t.Fatalf("NewSigstoreTufClient unexpectedly returned an error: %v", clientErr)
				}
				return
			}
			if tc.wantClientErr {
				t.Fatalf("NewSigstoreTufClient returned, expected error: %v", tc.wantClientErr)
			}
			initErr := client.Initialize(tc.repoOpts)
			if initErr != nil {
				if !tc.wantInitErr {
					t.Fatalf("Initialize unexpectedly returned an error: %v", initErr)
				}
				return
			}
			if tc.wantClientErr {
				t.Fatalf("Initialize returned, expected error: %v", tc.wantInitErr)
			}
		})
	}
}
