package httpx

import (
	"net/http"
	"os"
)

// Todo: Test performance / compare with standard `http.FileServer`

type noListingFilesystem struct {
	fs http.FileSystem
}

func (fs noListingFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return noReaddirFile{f}, nil
}

type noReaddirFile struct {
	http.File
}

func (f noReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func NoListingFileServer(root http.FileSystem) http.Handler {
	return http.FileServer(noListingFilesystem{root})
}
