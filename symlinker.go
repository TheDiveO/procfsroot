// Copyright 2023 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package procfsroot

import (
	"io/fs"
	"os"
)

// slinker defaults to using the os package functionality for reading and
// stat'ing symbolic links. It can be replaced when under test.
var slinker symlinker = &osSymlinker{}

// symlinker supports reading and stat'ing symbolic links.
type symlinker interface {
	Readlink(name string) (string, error)
	Lstat(name string) (fs.FileInfo, error)
}

// osSymlinker implements the symlinker interface using os stdlib functionality.
// We use it in production for direct, un-mocked access.
type osSymlinker struct{}

var _ symlinker = (*osSymlinker)(nil)

func (o *osSymlinker) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

func (o *osSymlinker) Lstat(name string) (fs.FileInfo, error) {
	return os.Lstat(name)
}
