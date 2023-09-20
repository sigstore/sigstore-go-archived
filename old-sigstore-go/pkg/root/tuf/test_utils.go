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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/theupdateframework/go-tuf"
)

type TestRepository struct {
	targets map[string]json.RawMessage
	td      string
	store   tuf.LocalStore
	repo    *tuf.Repo
	t       *testing.T
}

// newTufRepository initializes a TUF repository with root, targets, snapshot, and timestamp roles
func newTufRepository(t *testing.T, td string) *TestRepository {
	remote := tuf.FileSystemStore(td, nil)
	r, err := tuf.NewRepo(remote)
	if err != nil {
		t.Error(err)
	}
	if err := r.Init(false); err != nil {
		t.Error(err)
	}
	for _, role := range []string{"root", "targets", "snapshot", "timestamp"} {
		if _, err := r.GenKey(role); err != nil {
			t.Error(err)
		}
	}
	return &TestRepository{
		targets: make(map[string]json.RawMessage),
		td:      td,
		store:   remote,
		repo:    r,
		t:       t,
	}
}

func (r *TestRepository) addTarget(name string, data []byte, custom json.RawMessage) {
	targetPath := filepath.FromSlash(filepath.Join(r.td, "staged", "targets", name))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		r.t.Error(err)
	}
	if err := os.WriteFile(targetPath, data, 0o600); err != nil {
		r.t.Error(err)
	}
	if err := r.repo.AddTarget(name, custom); err != nil {
		r.t.Error(err)
	}
}

func (r *TestRepository) publish() {
	if err := r.repo.Snapshot(); err != nil {
		r.t.Error(err)
	}
	if err := r.repo.Timestamp(); err != nil {
		r.t.Error(err)
	}
	if err := r.repo.Commit(); err != nil {
		r.t.Error(err)
	}
}

func (r *TestRepository) root() []byte {
	meta, err := r.store.GetMeta()
	if err != nil {
		r.t.Error(err)
	}
	rootBytes, ok := meta["root.json"]
	if !ok {
		r.t.Error(err)
	}
	return rootBytes
}
