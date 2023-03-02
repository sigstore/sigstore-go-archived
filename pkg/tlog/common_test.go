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

package tlog

import (
	"encoding/hex"
	"testing"

	common_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	rekor_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
)

const rekor = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE2G2Y+2tabdTV5BcGiBIx0a9fAFwr
kBbmLSGtks4L3qX6yYY0zufBnhC8Ur/iy55GhWP/9A/bY2LhC30M9+RYtw==
-----END PUBLIC KEY-----`

func TestGetLogID(t *testing.T) {
	logID := []byte("foo")
	encodedLogID := hex.EncodeToString(logID)
	testCases := []struct {
		name       string
		entry      *rekor_v1.TransparencyLogEntry
		expectedID string
		wantErr    bool
	}{
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
			name: "fail: missing key id",
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex:          int64(1),
				LogId:             &common_v1.LogId{},
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: []byte("foo"),
				InclusionPromise:  &rekor_v1.InclusionPromise{},
			},
			wantErr: true,
		},
		{
			name: "valid: prod rekor",
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex: int64(1),
				LogId: &common_v1.LogId{
					KeyId: logID,
				},
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: []byte("foo"),
				InclusionPromise:  &rekor_v1.InclusionPromise{},
			},
			expectedID: encodedLogID,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			id, err := GetLogID(tc.entry)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("GetLogId unexpectedly returned an error: %v", err)
				}
				return
			}
			if tc.wantErr {
				t.Errorf("GetLogId returned, expected error")
				return
			}
			if tc.expectedID != id {
				t.Errorf("expected id %s, got %s", tc.expectedID, id)
			}
		})
	}
}

func TestComputeLogId(t *testing.T) {
	// Test the prod Rekor public key
	pubKey, err := cryptoutils.UnmarshalPEMToPublicKey([]byte(rekor))
	if err != nil {
		t.Error(err)
	}
	id, err := ComputeLogID(pubKey)
	if err != nil {
		t.Fatal(err)
	}
	expected := "c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d"
	if id != expected {
		t.Fatalf("expected %s, got %s", expected, id)
	}
}
