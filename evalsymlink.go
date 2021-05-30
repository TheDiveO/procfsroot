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
	"errors"
	"os"
	"path"
	"syscall"
)

// EvalSymlinkPathHandling tells EvalSymlinks how to behave with respect to the
// last path component when evaluating (chasing) symbolic links along a path.
type EvalSymlinkPathHandling = uint8

// Symlink evaluation of the last path component can either be enabled or
// disabled. All preceding path components will always be evaluated (resolved).
const (
	EvalFullPath   EvalSymlinkPathHandling = iota // evaluate all path components.
	EvalExceptLast                                // evaluate all but the last path component.
)

// EvalSymlinks returns the (absolute) path name after evaluating any symbolic
// links in the specified path, yet always with respect to a separately
// specified (and enforced) root. This function is modelled after Golang's
// standard library filepath.EvalSymlinks, but with the important differences
// that it only works on a unix-like filesystem and operates relative to an
// enforced root path.
//
// For instance, EvalSymlinks can be used to evaluate symlinks inside another
// mount namespace via the /proc/[PID]/root wormholes.
//
// The abspath parameter specifies an absolute path, which is taken relative
// (sic!) to the root parameter when evaluating symbolic links.
//
// The root parameter specifies the (enforced) root path when evaluating path
// components and symbolic links.
//
// The lastcomponent parameters specifies how to handle the final path component
// when chasing symbolic links: in some use cases the final path component
// should be taken as is instead of resolving it to the path it points to.
func EvalSymlinks(abspath, root string, pathhandling EvalSymlinkPathHandling) (string, error) {
	if len(abspath) < 1 || abspath[0] != '/' {
		abspath = "/" + abspath
	}
	dest := "/" // handle as absolute.
	jumps := 0
	var end int
	for start := 1; start < len(abspath); start = end {
		// Get the next path component by proceed to the beginning of the next
		// path component, skipping a "/" if necessary (including multiple
		// slashes as to be very forgiving here).
		for start < len(abspath) && abspath[start] == '/' {
			start++
		}
		// Now find the end of the path component, which actually will be the
		// following separator or *after* the end of the path.
		end = start
		for end < len(abspath) && abspath[end] != '/' {
			end++
		}
		// So, path[start:end] now is the path component for us to process,
		// without any slashes.
		if start == end {
			// no more path components to process, so we're done.
			break
		} else if abspath[start:end] == "." {
			// ignore and move on with the next path component, if any.
			continue
		} else if abspath[start:end] == ".." {
			// Try to remove the previous component (if there's one) from the
			// current dest.
			if dest == "/" {
				// Do not silently clamp to the root, but instead be rude by
				// throwing the towel.
				return "", errors.New("procfsroot.EvalSymlinks: no parent directory")
			}
			var prev int
			for prev = len(dest) - 1; prev > 0; prev-- {
				if dest[prev] == '/' {
					break
				}
			}
			if prev > 0 {
				dest = dest[:prev]
			} else {
				dest = "/"
			}
			continue
		}
		// This ain't no special component, so we can add it, trying avoiding
		// double slashes.
		if dest[len(dest)-1] != '/' {
			dest += "/"
		}
		dest += abspath[start:end]
		// We now need to find out more about this current destination: is it a
		// symbolic link we need to "evaluate", that is, to resolve by following
		// it? In case we're already arrived at the last component, the caller
		// might want us to not follow a final symlink, such as when working on
		// it itself instead of the referenced file system element? Or the final
		// element might not even exist as it is about to be created, but we
		// need to see where it will eventually land in the file system.
		if pathhandling == EvalExceptLast && end >= len(abspath) {
			break
		}
		wormpath := root + dest
		stat, err := os.Lstat(wormpath)
		if err != nil {
			return "", err
		}
		// For compatibility with pre-1.16 use "os" instead of "fs"
		if stat.Mode()&os.ModeSymlink == 0 {
			// Phew, ain't no symlink, so to ensure that -- with the exception
			// of the final component -- all other components are directories,
			// we check this path component if necessary.
			if !stat.Mode().IsDir() && end < len(abspath) {
				return "", syscall.ENOTDIR
			}
			continue
		}
		// It's a symbolic link and we put a simple guard in place (similar to
		// the ones from Jabberwocky) to break endless loops. After passing it
		// we try to make sense of where the symbolic link points to...
		jumps++
		if jumps > 255 {
			return "", errors.New("procfsroot.EvalSymlinks: too many symlinks")
		}
		link, err := os.Readlink(wormpath)
		if err != nil {
			return "", err
		}
		if link == "" {
			// An empty link has POSIX system-dependent behavior and with Linux
			// not being fully POSIX compliant, the system-dependent behavior
			// then is undefined. Oh well. Reading through "Empty symlinks and
			// full POSIX compliance" (https://lwn.net/Articles/551224/) we
			// rather play safe and simply refuse empty shambolic links, in case
			// they might be supported in some future.
			return "", errors.New("procfsroot.EvslSymlinks: rejecting empty symlink")
		} else if /* we KNOW it's not empty */ link[0] == '/' {
			// It's an absolute link, so we will continue with it. As the
			// absolute link may still sneak in ugly "." or "..", we cannot just
			// take it as our new destination but instead need to take it as our
			// new abspath and start all over again. Oh well.
			dest = "/"
			abspath = link + abspath[end:]
			end = 1 // restart ... doesn't really matter with or after slash.
		} else {
			// It's a relative link, so we need to drop the last path component
			// in the current dest and then make sure to process the link as
			// part of the abspath in order to correctly work on "." and "..".
			var prev int
			for prev = len(dest) - 1; prev > 0; prev-- {
				if dest[prev] == '/' {
					break
				}
			}
			if prev > 0 {
				dest = dest[:prev]
			} else {
				dest = "/"
			}
			abspath = link + abspath[end:]
			end = 0 // we prepended a relative path, so start "early".
		}
	}
	// Sanitize path to never climb back outside the worm hole, something we
	// definitely want to prohibit.
	final := path.Clean("/" + dest)
	return final, nil
}
