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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	common_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/common/v1"
	rekor_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore/pkg/signature"
)

const (
	encodedBody = "ayJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoicmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiJlY2RjNTUzNmY3M2JkYWU4ODE2ZjBlYTQwNzI2ZWY1ZTliODEwZDkxNDQ5MzA3NTkwM2JiOTA2MjNkOTdiMWQ4In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FWUNJUUQvUGRQUW1LV0MxKzBCTkVkNWdLdlFHcjF4eGwzaWVVZmZ2M2prMXp6Skt3SWhBTEJqM3hmQXlXeGx6NGpwb0lFSVYxVWZLOXZua1VVT1NvZVp4QlpQSEtQQyIsImZvcm1hdCI6Ing1MDkiLCJwdWJsaWNLZXkiOnsiY29udGVudCI6IkxTMHRMUzFDUlVkSlRpQlFWVUpNU1VNZ1MwVlpMUzB0TFMwS1RVWnJkMFYzV1VoTGIxcEplbW93UTBGUldVbExiMXBKZW1vd1JFRlJZMFJSWjBGRlRVOWpWR1pTUWxNNWFtbFlUVGd4UmxvNFoyMHZNU3R2YldWTmR3cHRiaTh6TkRjdk5UVTJaeTlzY21sVE56SjFUV2haT1V4alZDczFWVW8yWmtkQ1oyeHlOVm80VERCS1RsTjFZWE41WldRNVQzUmhVblozUFQwS0xTMHRMUzFGVGtRZ1VGVkNURWxESUV0RldTMHRMUzB0Q2c9PSJ9fX19"
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

	body, err := base64.RawStdEncoding.DecodeString(encodedBody)
	if err != nil {
		t.Fatalf("error decoding body: %v", err)
	}

	tlogEntry := &rekor_v1.TransparencyLogEntry{
		LogIndex: int64(1),
		LogId: &common_v1.LogId{
			Id: &common_v1.LogId_KeyId{
				KeyId: decodedLogID,
			},
		},
		IntegratedTime:    int64(1661794812),
		CanonicalizedBody: body,
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
			name: "fail: missing  entry log id",
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex:          int64(1),
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: body,
				InclusionPromise:  &rekor_v1.InclusionPromise{},
			},
			wantErr: true,
		},
		{
			name: "fail: missing entry promise",
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex: int64(1),
				LogId: &common_v1.LogId{
					Id: &common_v1.LogId_KeyId{
						KeyId: decodedLogID,
					},
				},
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: body,
			},
			wantErr: true,
		},
		{
			name:    "fail: invalid signature",
			wantErr: true,
			entry: &rekor_v1.TransparencyLogEntry{
				LogIndex: int64(1),
				LogId: &common_v1.LogId{
					Id: &common_v1.LogId_KeyId{
						KeyId: decodedLogID,
					},
				},
				IntegratedTime:    int64(1661794812),
				CanonicalizedBody: body,
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
