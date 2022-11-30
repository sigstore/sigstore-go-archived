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

/* This package implements rekor type generation and validation */
package types

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/go-openapi/runtime"
	common_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	rekor_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"

	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/rekor/pkg/types"
)

// GenerateTransparencyLogEntry creates a v1.TransparencylogEntry out
// of a Rekor Log Entry. We assume the Rekor entry is well-formed (e.g.
// the log ID fields are valid log IDs, the kind and version are valid).
func GenerateTransparencyLogEntry(entry *models.LogEntryAnon) (
	*rekor_v1.TransparencyLogEntry, error) {
	log_id, err := hex.DecodeString(*entry.LogID)
	if err != nil {
		return nil, fmt.Errorf("decoding log ID: %w", err)
	}
	set, err := base64.StdEncoding.DecodeString(string(entry.Verification.SignedEntryTimestamp))
	if err != nil {
		return nil, fmt.Errorf("decoding SignedEntryTimestamp: %w", err)
	}
	body, err := base64.StdEncoding.DecodeString(entry.Body.(string))
	if err != nil {
		return nil, fmt.Errorf("decoding Rekor entry body: %w", err)
	}
	entryKindVersion, err := entryToKindVersion(entry.Body.(string))
	if err != nil {
		return nil, fmt.Errorf("entryToKindVersion: %w", err)
	}
	res := &rekor_v1.TransparencyLogEntry{
		LogIndex: *entry.LogIndex,
		LogId: &common_v1.LogId{
			Id: &common_v1.LogId_KeyId{
				KeyId: log_id,
			}},
		KindVersion:    entryKindVersion,
		IntegratedTime: *entry.IntegratedTime,
		InclusionPromise: &rekor_v1.InclusionPromise{
			SignedEntryTimestamp: set,
		},
		CanonicalizedBody: body,
	}
	return res, nil
}

// entryToKindVersion extracts the Kind and Version out of a Rekor entry body.
func entryToKindVersion(e string) (*rekor_v1.KindVersion, error) {
	res := &rekor_v1.KindVersion{}
	b, err := base64.StdEncoding.DecodeString(e)
	if err != nil {
		return nil, err
	}
	pe, err := models.UnmarshalProposedEntry(bytes.NewReader(b), runtime.JSONConsumer())
	if err != nil {
		return nil, err
	}
	res.Kind = pe.Kind()
	entry, err := types.UnmarshalEntry(pe)
	if err != nil {
		return nil, err
	}
	res.Version = entry.APIVersion()
	return res, nil
}
