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
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStoreFromOpts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	created := filepath.Join(dir, "created")
	if err := os.Mkdir(created, 0750); err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		name      string
		opts      *TUFClientOptions
		wantError error
	}{
		{
			name: "unkown cache type",
			opts: &TUFClientOptions{
				CacheType: 3,
			},
			wantError: errUnknownCacheType,
		},
		{
			name: "cache type memory",
			opts: &TUFClientOptions{
				CacheType: Memory,
			},
		},
		{
			name: "cache type disk no location",
			opts: &TUFClientOptions{
				CacheType: Disk,
			},
			wantError: errUnknownCacheLocation,
		},
		{
			name: "cache type disk new cache",
			opts: &TUFClientOptions{
				CacheType:     Disk,
				CacheLocation: filepath.Join(dir, "test"),
			},
		},
		{
			name: "cache type disk already created",
			opts: &TUFClientOptions{
				CacheType:     Disk,
				CacheLocation: created,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := localStoreFromOpts(tc.opts)
			if err != nil {
				if tc.wantError == nil {
					t.Fatalf("localStoreFromOpts unexpectedly returned an error: %v", err)
				}
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("localStoreFromOpts returned %v, expected %v", err, tc.wantError)
				}
				return
			}
			if tc.wantError != nil {
				t.Errorf("localStoreFromOpts returned, expected error: %v", err)
			}
		})
	}
}

func TestRemoteStoreFromOpts(t *testing.T) {
	fileRemote := t.TempDir()
	if err := os.Mkdir(filepath.Join(fileRemote, "targets"), 0750); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		opts      *RepositoryOptions
		wantError bool
	}{
		{
			name: "local file remote",
			opts: &RepositoryOptions{
				Remote: fmt.Sprintf("file://%s", fileRemote),
			},
		},
		{
			name: "local file remote bad URI",
			opts: &RepositoryOptions{
				Remote: "abc",
			},
			wantError: true,
		},
		{
			name: "local file remote does not exist",
			opts: &RepositoryOptions{
				Remote: "file://abc",
			},
			wantError: true,
		},
		{
			name: "http file remote",
			opts: &RepositoryOptions{
				Remote: "http://abc",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := remoteStoreFromOpts(tc.opts)
			if err != nil {
				if !tc.wantError {
					t.Fatalf("remoteStoreFromOpts unexpectedly returned an error: %v", err)
				}
				return
			}
			if tc.wantError {
				t.Errorf("remoteStoreFromOpts returned, expected error: %v", err)
			}
		})
	}
}
