package fileutil

import (
    "fmt"
    "io"
    "os"
    "strings"
)

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
    sfi, err := os.Stat(src)
    if err != nil {
        return
    }
    if !sfi.Mode().IsRegular() {
        // cannot copy non-regular files (e.g., directories,
        // symlinks, devices, etc.)
        return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
    }
    dfi, err := os.Stat(dst)
    if err != nil {
        if !os.IsNotExist(err) {
            // if there's an error (other than file doesn't exist)
            return
        }
    } else {
        if !(dfi.Mode().IsRegular()) {
            return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
        }
        if os.SameFile(sfi, dfi) {
            return
        }
    }
    if err = os.Link(src, dst); err == nil {
        return
    }
    err = copyFileContents(src, dst)
    return
}

func ExistFile(file string) (bool) {
	if _, err := os.Stat(file); err == nil {
		return true
	} else {
		return false
	}
}

// SymlinkFile creats link (ln) to a file (tgt). If tgt and ln files exist, and are
// the same, then return success. Otherwise, attempt to create or overwrite a symlink
// between the two files.
func SymlinkFile(tgt, ln string) (err error) {
	pwd, _ := os.Getwd()
	if ! strings.HasPrefix(tgt, "/") {
        tgt = pwd + "/" + tgt
    }
	sfi, err := os.Stat(tgt)
	if err != nil {
		return err
	}
	if !sfi.Mode().IsRegular() {
		// cannot link to non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("SymlinkFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
    dfi, err := os.Lstat(ln)
    if err == nil {
        // So file / link exists!
        // Check if it's not a regular file and not a symlink
		if !dfi.Mode().IsRegular() {
            // Not regular file
            if dfi.Mode() & os.ModeSymlink == 0 {
                // not a symlink also...
                return fmt.Errorf("SymlinkFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
            }
		}
		if os.SameFile(sfi, dfi) {
			return nil
		}
        if (dfi.Mode() & os.ModeSymlink != 0) {
            var existingLink string
            existingLink, err = os.Readlink(ln)
            if existingLink == tgt {
                // No change required or err
                return nil
            } else {
                // Replace link
                err = os.Remove(ln)
                if err != nil {
                    return err
                }
            }
        } else {
            return fmt.Errorf("SymlinkFile: not replacing existing (non-symlink) file %s", ln)
        }
	}
	err = os.Symlink(tgt, ln)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()
    return
}