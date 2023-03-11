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

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const fsroot = "./test/root"

var _ = Describe("evil symlink chasing", func() {

	It("handles simple paths", func() {
		p, err := EvalSymlinks("/a/b.txt", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a/b.txt"))

		p, err = EvalSymlinks("/////a/////b.txt", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a/b.txt"))

		p, err = EvalSymlinks("", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/"))

		p, err = EvalSymlinks("a", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a"))

		Expect(EvalSymlinks("/a/b.txt/c", fsroot, EvalFullPath)).
			Error().To(HaveOccurred())

		Expect(EvalSymlinks("/a/zzz/b.txt", fsroot, EvalFullPath)).
			Error().To(HaveOccurred())

		p, err = EvalSymlinks("//a//", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a"))
	})

	It("handles . and ..", func() {
		p, err := EvalSymlinks("/./a/./b.txt", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a/b.txt"))

		p, err = EvalSymlinks("/a/../a/b.txt", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a/b.txt"))

		p, err = EvalSymlinks("/a/d/../b.txt", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a/b.txt"))
	})

	It("expects file path elements to exist", func() {
		Expect(EvalSymlinks("/a/zzz/whateverelse", fsroot, EvalFullPath)).
			Error().To(HaveOccurred())
	})

	It("optionally accepts missing target", func() {
		Expect(EvalSymlinks("/a/zzz.txt", fsroot, EvalFullPath)).
			Error().To(HaveOccurred())

		p, err := EvalSymlinks("/a/zzz.txt", fsroot, EvalExceptLast)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a/zzz.txt"))
	})

	It("follows symlinks", func() {
		p, err := EvalSymlinks("/relsymlink", fsroot, EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/a/b.txt"))

		p, err = EvalSymlinks("/var/run", "/", EvalFullPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(p).To(Equal("/run"))

		Expect(EvalSymlinks("/proc/foobar1/root", "/", EvalFullPath)).
			Error().To(HaveOccurred())
	})

	It("stays inside the wormhole", func() {
		Expect(EvalSymlinks("/../foo", fsroot, EvalFullPath)).
			Error().To(MatchError(ContainSubstring("no parent directory")))

		Expect(EvalSymlinks("/a/d/../../../foo", fsroot, EvalFullPath)).
			Error().To(MatchError(ContainSubstring("no parent directory")))

		Expect(EvalSymlinks("/unrooter/tryingtoleavethebox", fsroot, EvalFullPath)).
			Error().To(MatchError(ContainSubstring("no parent directory")))
	})

	It("doesn't follow endlessly", func() {
		ouroboros := strings.Repeat("/proc/self/root", 256*2)
		Expect(EvalSymlinks(ouroboros, "/proc/self/root", EvalFullPath)).
			Error().To(MatchError(ContainSubstring("too many symlinks")))
	})

})
