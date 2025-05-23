package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type PeopleFS struct {
	mu          sync.RWMutex
	peopleIndex PeopleIndex
	photoIndex  PhotoIndex
}

func NewPeopleFS() *PeopleFS {
	return &PeopleFS{}
}

func (p *PeopleFS) Reindex(inputDir string) error {
	newPhotoIndex, newPeopleIndex, err := extractPeopleTags(inputDir)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.peopleIndex = newPeopleIndex
	p.photoIndex = newPhotoIndex
	p.mu.Unlock()
	log.Println("Index updated")
	return nil
}

func (p *PeopleFS) Root() (fs.Node, error) {
	return &Dir{fs: p, path: "/"}, nil
}

type Dir struct {
	fs   *PeopleFS
	path string
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.fs.mu.RLock()
	defer d.fs.mu.RUnlock()
	entries := []fuse.Dirent{}
	if d.path == "/" {
		// At root level, list the people directories.
		for person, _ := range d.fs.peopleIndex {
			entries = append(entries, fuse.Dirent{Name: person, Type: fuse.DT_Dir})
		}
		return entries, nil
	} else {
		dirs := strings.SplitN(strings.TrimPrefix(d.path, "/"), "/", 2)
		if files := d.fs.peopleIndex[dirs[0]]; files != nil {

			addedDirs := make([]string, 0)
			for _, file := range files {
				var a string
				if len(dirs) > 1 {
					if !strings.HasPrefix(file, inputDir+dirs[1]+"/") {
						continue
					}
					a = strings.TrimPrefix(file, inputDir+dirs[1]+"/")
				} else {
					if !strings.HasPrefix(file, inputDir) {
						continue
					}
					a = strings.TrimPrefix(file, inputDir)
				}
				subdirs := strings.SplitN(a, "/", 3)
				if len(subdirs) > 1 {
					alreadyAdded := false
					for _, dir := range addedDirs {
						if strings.HasPrefix(subdirs[0], dir) {
							alreadyAdded = true
							break
						}
					}
					if !alreadyAdded {
						addedDirs = append(addedDirs, subdirs[0])
						entries = append(entries, fuse.Dirent{Name: subdirs[0], Type: fuse.DT_Dir})
					}
				} else {
					entries = append(entries, fuse.Dirent{Name: subdirs[0], Type: fuse.DT_File})
				}
				// entries = append(entries, fuse.Dirent{Name: filepath.Base(file), Type: fuse.DT_File})
			}
			return entries, nil
		} else {
			// If the directory is not found in the index, return an empty list.
			return []fuse.Dirent{}, nil
		}
	}

	return []fuse.Dirent{}, nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	full := filepath.Join(d.path, name)
	d.fs.mu.RLock()
	defer d.fs.mu.RUnlock()

	// Subdirectory
	if d.path == "/" {
		return &Dir{fs: d.fs, path: full}, nil
	}

	// d.path = strings.TrimSuffix(d.path, "/")
	path := strings.TrimPrefix(full, "/")
	dirs := strings.SplitN(path, "/", 2)
	//files := d.fs.peopleIndex[dirs[0]]
	people := d.fs.photoIndex[inputDir+dirs[1]]
	if people == nil {

		return &Dir{fs: d.fs, path: full}, nil
	}
	for _, name := range people {
		if dirs[0] == name {
			return &File{path: inputDir + dirs[1]}, nil
		}
	}
	return nil, fuse.ENOENT
}

type File struct {
	path string
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	st, err := os.Stat(f.path)
	if err != nil {
		return fuse.ENOENT
	}
	a.Mode = st.Mode()
	a.Size = uint64(st.Size())
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	return os.ReadFile(f.path)
}

func MountFS(mountPoint string, pfs *PeopleFS) error {
	conn, err := fuse.Mount(mountPoint, fuse.ReadOnly(), fuse.AllowOther())
	if err != nil {
		return err
	}
	defer conn.Close()
	return fs.Serve(conn, pfs)
}
func UnmountFS(mountPoint string) error {
	cmd := exec.Command("fusermount", "-u", mountPoint)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to unmount filesystem: %v", err)
	}
	return nil
}
