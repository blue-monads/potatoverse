package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/psanford/sqlite3vfs"
)

var _ sqlite3vfs.VFS = (*VFSync)(nil)

func init() {
	sqlite3vfs.RegisterVFS("vfsync", NewVFSync("./data"))
}

type VFSync struct {
	baseDir    string
	absBaseDir string
}

func NewVFSync(baseDir string) *VFSync {
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		panic(err)
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		panic(err)
	}

	return &VFSync{
		baseDir:    baseDir,
		absBaseDir: absBaseDir,
	}
}

// isPathSafe checks if the given path is within the base directory
func (vfs *VFSync) isPathSafe(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(vfs.absBaseDir, absPath)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func (vfs *VFSync) Open(name string, flags sqlite3vfs.OpenFlag) (sqlite3vfs.File, sqlite3vfs.OpenFlag, error) {
	var (
		f   *os.File
		err error
	)

	qq.Println("Open", name, flags)

	var fname string
	if filepath.IsAbs(name) {
		// If name is already absolute (from FullPathname), use it directly
		fname = name
	} else {
		// Otherwise, join with baseDir
		fname = filepath.Join(vfs.baseDir, name)
	}

	if !vfs.isPathSafe(fname) {
		return nil, 0, sqlite3vfs.PermError
	}

	var fileFlags int
	if flags&sqlite3vfs.OpenExclusive != 0 {
		fileFlags |= os.O_EXCL
	}
	if flags&sqlite3vfs.OpenCreate != 0 {
		fileFlags |= os.O_CREATE
	}
	if flags&sqlite3vfs.OpenReadOnly != 0 {
		fileFlags |= os.O_RDONLY
	}
	if flags&sqlite3vfs.OpenReadWrite != 0 {
		fileFlags |= os.O_RDWR
	}

	if flags&sqlite3vfs.OpenWAL != 0 {
		fileFlags |= os.O_APPEND
	}

	if fileFlags == 0 {
		fileFlags = os.O_RDWR
	}

	f, err = os.OpenFile(fname, fileFlags, 0600)
	if err != nil {
		return nil, 0, sqlite3vfs.CantOpenError
	}

	tf := &VFSyncFile{f: f}
	return tf, flags, nil
}

func (vfs *VFSync) Delete(name string, dirSync bool) error {
	qq.Println("Delete", name, dirSync)

	var fname string
	if filepath.IsAbs(name) {
		qq.Println("Delete/0", name)
		fname = name
	} else {
		qq.Println("Delete/1", vfs.baseDir, name)
		fname = filepath.Join(vfs.baseDir, name)
	}
	if !vfs.isPathSafe(fname) {
		qq.Println("Delete/2", fname)
		return errors.New("illegal path")
	}

	qq.Println("Delete/3", fname)
	//return os.Remove(fname)
	// rename the file to .deleted

	randID := rand.Intn(1000000)
	deletedFile := fmt.Sprintf("%s.%d.deleted", fname, randID)
	err := os.Rename(fname, deletedFile)
	if err != nil {
		qq.Println("Delete/4", err)
		return err
	}

	qq.Println("Delete/5", deletedFile)
	return nil
}

func (vfs *VFSync) Access(name string, flag sqlite3vfs.AccessFlag) (bool, error) {

	qq.Println("Access", name, flag)

	var fname string
	if filepath.IsAbs(name) {
		fname = name
	} else {
		fname = filepath.Join(vfs.baseDir, name)
	}
	if !vfs.isPathSafe(fname) {
		return false, errors.New("illegal path")
	}

	exists := true
	_, err := os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			exists = false
		} else {
			// For permission errors or other stat errors, treat as non-existent
			// This is safer and matches SQLite's expectations
			exists = false
		}
	}

	if flag == sqlite3vfs.AccessExists {
		return exists, nil
	}

	return true, nil
}

func (vfs *VFSync) FullPathname(name string) string {
	qq.Println("FullPathname", name)

	// Return the absolute path that Open can use directly
	fname := filepath.Join(vfs.baseDir, name)
	absPath, err := filepath.Abs(fname)
	if err != nil {
		// Fallback to relative path if Abs fails
		return fname
	}
	return absPath
}

type VFSyncFile struct {
	lockCount int64
	f         *os.File
}

func (tf *VFSyncFile) Close() error {
	qq.Println("Close")
	return tf.f.Close()
}

func (tf *VFSyncFile) ReadAt(p []byte, off int64) (n int, err error) {
	qq.Println("ReadAt", len(p), off)
	return tf.f.ReadAt(p, off)
}

func (tf *VFSyncFile) WriteAt(b []byte, off int64) (n int, err error) {
	qq.Println("WriteAt", len(b), off)
	return tf.f.WriteAt(b, off)
}

func (tf *VFSyncFile) Truncate(size int64) error {
	qq.Println("Truncate", size)
	return tf.f.Truncate(size)
}

func (tf *VFSyncFile) Sync(flag sqlite3vfs.SyncType) error {
	qq.Println("Sync", flag)
	return tf.f.Sync()
}

func (tf *VFSyncFile) FileSize() (int64, error) {
	qq.Println("FileSize")

	cur, _ := tf.f.Seek(0, io.SeekCurrent)
	end, err := tf.f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	tf.f.Seek(cur, io.SeekStart)
	return end, nil
}

func (tf *VFSyncFile) Lock(elock sqlite3vfs.LockType) error {
	qq.Println("Lock", elock)

	if elock == sqlite3vfs.LockNone {
		return nil
	}
	atomic.AddInt64(&tf.lockCount, 1)
	return nil
}

func (tf *VFSyncFile) Unlock(elock sqlite3vfs.LockType) error {
	qq.Println("Unlock", elock)

	if elock == sqlite3vfs.LockNone {
		return nil
	}
	atomic.AddInt64(&tf.lockCount, -1)
	return nil
}

func (tf *VFSyncFile) CheckReservedLock() (bool, error) {
	qq.Println("CheckReservedLock")

	count := atomic.LoadInt64(&tf.lockCount)
	return count > 0, nil
}

func (tf *VFSyncFile) SectorSize() int64 {
	qq.Println("SectorSize")
	return 0
}

func (tf *VFSyncFile) DeviceCharacteristics() sqlite3vfs.DeviceCharacteristic {
	qq.Println("DeviceCharacteristics")
	// Use DefaultDeviceCharacteristics for a reasonable set of capabilities
	// that work on most modern filesystems (ext4, xfs, APFS, etc.)
	// This includes:
	//   - Atomic writes up to 4KB (IocapAtomic4K, IocapAtomic2K, IocapAtomic1K, IocapAtomic512)
	//   - Safe append (IocapSafeAppend)
	//   - Power-safe overwrites (IocapPowersafeOverwrite)
	// return sqlite3vfs.DefaultDeviceCharacteristics()
	return 0
}

/*

func (tf *VFSyncFile) ShmMap(iPg int, pgsz int, isWrite bool) ([]byte, error) {
	qq.Println("ShmMap", iPg, pgsz, isWrite)
	// Return NotFoundError to indicate shared memory is not supported
	return nil, sqlite3vfs.NotFoundError
}

func (tf *VFSyncFile) ShmLock(offset int, n int, flags sqlite3vfs.ShmLockFlag) error {
	qq.Println("ShmLock", offset, n, flags)
	// Return NotFoundError to indicate shared memory is not supported
	return sqlite3vfs.NotFoundError
}

func (tf *VFSyncFile) ShmBarrier() {
	qq.Println("ShmBarrier")
	// No-op for files that don't support shared memory
}

func (tf *VFSyncFile) ShmUnmap(deleteFlag bool) error {
	qq.Println("ShmUnmap", deleteFlag)
	// Return NotFoundError to indicate shared memory is not supported
	return sqlite3vfs.NotFoundError
}

*/
