/*

Package procfsroot helps with accessing file system paths containing absolute
symbolic links that are to be taken relative (sic!) to a particular root path. A
good example is accessing paths inside /proc/[PID]/root "wormholes" in the proc
file system. Symbolic links are properly resolved and kept inside a given root
path, prohibiting rogue relative symbolic links from breaking out of a procfs
root wormhole.

*/
package procfsroot
