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

// This package implements a root provider interface that can be
//
//	implemented with a TUF client or other
package root

// TrustedRootProvider is an interface that can generate a trusted
// root, be it from a TUF client, local filesystem information, or
// other method to retrieve the trusted root.
type TrustedRootProvider interface {
	// GetTrustedRoot returns a TrustedRootStore containing the
	// Sigstore ecosystem information for a verification client to
	// consume.
	GetTrustedRoot() (interface{}, error)
}
