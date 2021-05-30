// Copyright 2021 Harald Albrecht.
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

import "fmt"

// Evaluating symlinks returns an absolute path (interpreted relative to the
// root path) with ".", "..", and all symbolic links properly resolved. The
// returned path does not contain the prefixing root path.
func ExampleEvalSymlinks() {
	root := "/proc/self/root"
	path, err := EvalSymlinks("/var/run/docker.sock", root, EvalFullPath)
	if err != nil {
		panic(err)
	}
	fmt.Println(path)
	// Output: /run/docker.sock
}

// Evaluating a path with non-existing path components returns an error.
func ExampleEvalSymlinks_error() {
	root := "/proc/self/root"
	_, err := EvalSymlinks("/var/run/something", root, EvalFullPath)
	fmt.Println(err)
	// Output: lstat /proc/self/root/run/something: no such file or directory
}

// In some use cases it might be necessary not to resolve the final path
// component in order to be able to work on a trailing symbolic link component
// itself instead of where the symlink points to.
func ExampleEvalSymlinks_evalexceptlast() {
	root := "/proc/self/root"
	p, err := EvalSymlinks("/var/run/something", root, EvalExceptLast)
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
	// Output: /run/something
}

// Evaluated paths cannot break out of their root path "sandbox", neither using
// ".." nor relative symbolic links containing enough "..".
func ExampleEvalSymlinks_boxed() {
	root := "/proc/self/root"
	_, err := EvalSymlinks("/../../../bin/hostbinary", root, EvalFullPath)
	fmt.Println(err)
	// Output: procfsroot.EvalSymlinks: no parent directory
}
