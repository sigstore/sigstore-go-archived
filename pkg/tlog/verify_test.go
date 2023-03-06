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

/*This package implements tlog verification functions */
package tlog

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	common_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	rekor_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore/pkg/signature"
)

func TestVerifyTlogSET(t *testing.T) {
	t.Parallel()

	signer, _, err := signature.NewDefaultECDSASignerVerifier()
	if err != nil {
		t.Fatalf("error generating signer: %v", err)
	}

	logID, err := ComputeLogID(signer.Public())
	if err != nil {
		t.Fatalf("getting log id: %v", err)
	}
	decodedLogID, err := hex.DecodeString(logID)
	if err != nil {
		t.Fatalf("decoding log id: %v", err)
	}

	tlogEntry := &rekor_v1.TransparencyLogEntry{
		LogIndex: int64(1),
		LogId: &common_v1.LogId{
			KeyId: decodedLogID,
		},
		IntegratedTime:    int64(1661794812),
		CanonicalizedBody: []byte("foo"),
		InclusionPromise:  &rekor_v1.InclusionPromise{},
	}
	payload, err := verificationPayload(tlogEntry)
	if err != nil {
		t.Fatal(err)
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	canonicalized, err := jsoncanonicalizer.Transform(jsonPayload)
	if err != nil {
		t.Fatal(err)
	}
	tlogEntry.InclusionPromise.SignedEntryTimestamp, err = signer.SignMessage(bytes.NewReader(canonicalized))
	if err != nil {
		t.Fatal(err)
	}

	trustedKey := map[string]signature.Verifier{
		logID: signer,
	}

	testCases := []struct {
		name    string
		entry   *rekor_v1.TransparencyLogEntry
		keys    map[string]signature.Verifier
		wantErr bool
	}{
		{
			name:  "valid: valid tlog set",
			keys:  trustedKey,
			entry: tlogEntry,
		},
		{
			name:    "fail: missing trusted tlog key",
			entry:   tlogEntry,
			keys:    map[string]signature.Verifier{},
			wantErr: true,
		},
		{
			name: "fail: missing entry log id",
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex:          int64(1),
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: []byte("foo"),
				InclusionPromise:  &rekor_v1.InclusionPromise{},
			},
			wantErr: true,
		},
		{
			name: "fail: missing entry promise",
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex: int64(1),
				LogId: &common_v1.LogId{
					KeyId: decodedLogID,
				},
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: []byte("foo"),
			},
			wantErr: true,
		},
		{
			name:    "fail: invalid signature",
			wantErr: true,
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex: int64(1),
				LogId: &common_v1.LogId{
					KeyId: decodedLogID,
				},
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: []byte("foo"),
				InclusionPromise: &rekor_v1.InclusionPromise{
					SignedEntryTimestamp: []byte("foo"),
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			err := VerifyTlogSET(ctx, tc.entry, tc.keys)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("VerifyTlogSET unexpectedly returned an error: %v", err)
				}
				return
			}
			if tc.wantErr {
				t.Errorf("VerifyTlogSET returned, expected error")
			}
		})
	}
}
