package fixtures

import (
	"errors"
	"os"
)

func CreateFileLoader(location string) ContentLoader {
	return &contentLoaderImpl{
		location: location,
		known:    map[string]bool{},
		suffixes: []string{"", ".yml", ".yaml"},
	}
}

type contentLoaderImpl struct {
	location string
	known    map[string]bool
	suffixes []string
}

func (l *contentLoaderImpl) Load(name string) (string, []byte, error) {
	file, err := l.findFixturePath(l.location, name)
	if err != nil {
		return "", nil, err
	}

	// skip previously loaded files
	if l.known[file] {
		return file, []byte{}, nil
	}
	l.known[file] = true

	data, err := os.ReadFile(file)
	if err != nil {
		return "", nil, err
	}
	return file, data, nil
}

func (l *contentLoaderImpl) findFixturePath(location, name string) (string, error) {
	var err error
	for _, suffix := range l.suffixes {
		candidate := location + "/" + name + suffix
		var stats os.FileInfo
		if stats, err = os.Stat(candidate); err == nil && !stats.IsDir() {
			return candidate, nil
		}
	}
	if os.IsNotExist(err) {
		return "", errors.New("file not found")
	}
	return "", err
}
