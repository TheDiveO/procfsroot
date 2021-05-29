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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/errxpect"
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

		Errxpect(EvalSymlinks("/a/b.txt/c", fsroot, EvalFullPath)).
			To(HaveOccurred())

		Errxpect(EvalSymlinks("/a/zzz/b.txt", fsroot, EvalFullPath)).
			To(HaveOccurred())

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
		Errxpect(EvalSymlinks("/a/zzz/whateverelse", fsroot, EvalFullPath)).
			To(HaveOccurred())
	})

	It("optionally accepts missing target", func() {
		Errxpect(EvalSymlinks("/a/zzz.txt", fsroot, EvalFullPath)).
			To(HaveOccurred())

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

		Errxpect(EvalSymlinks("/proc/1/root", "/", EvalFullPath)).
			To(HaveOccurred())
	})

	It("stays inside the wormhole", func() {
		Errxpect(EvalSymlinks("/../foo", fsroot, EvalFullPath)).
			To(MatchError(ContainSubstring("no parent directory")))

		Errxpect(EvalSymlinks("/a/d/../../../foo", fsroot, EvalFullPath)).
			To(MatchError(ContainSubstring("no parent directory")))

		Errxpect(EvalSymlinks("/unrooter/tryingtoleavethebox", fsroot, EvalFullPath)).
			To(MatchError(ContainSubstring("no parent directory")))
	})

	It("doesn't follow endlessly", func() {
		ouroboros := strings.Repeat("/proc/self/root", 256*2)
		Errxpect(EvalSymlinks(ouroboros, "/proc/self/root", EvalFullPath)).
			To(MatchError(ContainSubstring("too many symlinks")))
	})

})
