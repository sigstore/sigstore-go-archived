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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	rekor_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
	"github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/signature/options"
)

// VerificationPayload is a struct containing the payload the
// SignedEntryTimestamp signs over. The signed payload is constructed from
// the JSON canonicalized bytes of this struct.
type VerificationPayload struct {
	Body           interface{} `json:"body"`
	IntegratedTime int64       `json:"integratedTime"`
	LogIndex       int64       `json:"logIndex"`
	LogID          string      `json:"logID"`
}

// VerifyTlogSET verifies the SignedEntryTimestamp (SET) for the given
// TransparencyLogEntry using the trusted verifiers indexed by LogID.
func VerifyTlogSET(ctx context.Context,
	entry *rekor_v1.TransparencyLogEntry, trustedKeys map[string]signature.Verifier,
) error {
	// Create the signed tlog verification payload.
	payload, err := verificationPayload(entry)
	if err != nil {
		return fmt.Errorf("creating verification payload: %w", err)
	}

	// Canonicalize using JSON canonicalizer
	contents, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling : %w", err)
	}
	canonicalized, err := jsoncanonicalizer.Transform(contents)
	if err != nil {
		return fmt.Errorf("canonicalizing: %w", err)
	}

	// Find the corresponding public key which generated the SET
	// We should already have a LogId, or else the verification payload would fail.
	entryLogID, err := GetLogID(entry)
	if err != nil {
		return fmt.Errorf("getting entry log ID: %w", err)
	}
	verifier, ok := trustedKeys[entryLogID]
	if !ok {
		return errors.New("rekor log public key not found for payload")
	}

	// Extract the SET from the tlog entry
	if entry.GetInclusionPromise() == nil {
		return errors.New("rekor entry missing inclusion promise")
	}
	sig := entry.InclusionPromise.SignedEntryTimestamp

	// Verify the SET over the payload
	if err := verifier.VerifySignature(bytes.NewReader(sig),
		bytes.NewReader(canonicalized), options.WithContext(ctx)); err != nil {
		return fmt.Errorf("unable to verify bundle: %w", err)
	}

	return nil
}

func verificationPayload(entry *rekor_v1.TransparencyLogEntry) (*VerificationPayload, error) {
	if entry.GetLogId() == nil {
		return nil, errors.New("TransparencyLogEntry missing LogId")
	}

	return &VerificationPayload{
		Body:           base64.StdEncoding.EncodeToString(entry.CanonicalizedBody),
		IntegratedTime: entry.IntegratedTime,
		LogIndex:       entry.LogIndex,
		LogID:          entry.LogId.String(),
	}, nil
}
