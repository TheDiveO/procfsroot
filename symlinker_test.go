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

	"github.com/spf13/afero"
)

// aferoSymlinker adapts an afero filesystem to our symlinker interface. For
// this, the specific afero Fs implementation must implement the afero
// LinkReader and Lstater interfaces.
type aferoSymlinker struct {
	afero.Fs
}

var _ symlinker = (*aferoSymlinker)(nil)

func (a aferoSymlinker) Readlink(name string) (string, error) {
	return a.Fs.(afero.LinkReader).ReadlinkIfPossible(name)
}

func (a aferoSymlinker) Lstat(name string) (fs.FileInfo, error) {
	fi, _, err := a.Fs.(afero.Lstater).LstatIfPossible(name)
	return fi, err
}
