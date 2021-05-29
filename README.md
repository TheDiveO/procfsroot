# ProcfsRoot

`procfsroot` is a small Go module that helps with accessing file system paths
containing absolute symbolic links that are to be taken relative (sic!) to a
particular root path. A good example is accessing paths inside
`/proc/[PID]/root` "wormholes" in the [proc file
system](https://man7.org/linux/man-pages/man5/proc.5.html). Symbolic links are
properly resolved and kept inside a given root path, prohibiting rogue relative
symbolic links from breaking out of a procfs root wormhole.

## Usage

`procfsroot.EvalSymlinks()` mirrors Golang's
[`filepath.EvalSymlinks`](https://golang.org/pkg/path/filepath/#EvalSymlinks),
but works only on paths using "`/`" forward slashes and enforces symbolic link
chasing relative to an enforced root path.

In the following example, the "absolute" path `/var/run/docker.sock` (which
might be in a different mount namespace) is correctly resolved in the root
context of `/proc/1/root` – even in case of absolute symbolic links. Trying to
directly use `/proc/1/root/var/run/docker.sock` will fail in case of different
mount namespaces between the accessing process and the initial mount namespace
of the init process PID 1.

```go
import (
    "os"
    "github.com/thediveo/procfsroot"
)

var f, err := os.Open(
    procfsroot.EvalSymlinks("/var/run/docker.sock", "/proc/1/root", procfsroot.EvalFullPath))
```

## Mount Namespace Wormholes

In case you have either never noticed the special `/proc/[PID]/root` links or
have ever wondered what they're good for: they're kind of "wormholes" into
arbitrary [mount
namespaces](https://man7.org/linux/man-pages/man7/mount_namespaces.7.html) given
a suitable process ID (PID). They simplify accessing directories and files in
other mount namespaces because they do not require switching the accessing
process first into the target mount namespace (which can only be done while
single threaded).

| Access Method | Required Capabilites |
| :--- | :--- |
| `setns()` | `CAP_SYS_ADMIN`, `CAP_SYS_CHROOT`, as well as typically also `CAP_SYS_PTRACE` in order to access a mount namespace reference in `/proc/[PID]/ns/mnt`</li></ul> |
| `/proc/[PID]/root` | `CAP_SYS_PTRACE` (so convenient 😀) |

Of course, the usual file system DAC (discretionary access control) still
applies as usual – including UID 0 access rules.

Also, for access to `/proc/[PID]` the current process needs to be in a suitable
[PID namespace](https://man7.org/linux/man-pages/man7/pid_namespaces.7.html)
that includes the PID of a "target" process of interest. Of course, the initial
PID namespace is "gold standard".

## Copyright and License

Copyright 2021 Harald Albrecht, licensed under the Apache License, Version 2.0.
