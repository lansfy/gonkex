package sqldb

import (
	"errors"
	"os"
)

type ContentLoader interface {
	Load(name string) (string, []byte, error)
}

func CreateFileLoader(location string) ContentLoader {
	return &contentLoaderImpl{
		location: location,
		known:    map[string]bool{},
	}
}

type contentLoaderImpl struct {
	location string
	known    map[string]bool
}

func (l *contentLoaderImpl) Load(name string) (string, []byte, error) {
	file, err := findFixturePath(l.location, name)
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

func findFixturePath(location, name string) (string, error) {
	candidates := []string{
		location + "/" + name,
		location + "/" + name + ".yml",
		location + "/" + name + ".yaml",
	}

	var err error
	for _, candidate := range candidates {
		if _, err = os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	if os.IsNotExist(err) {
		return "", errors.New("file not exists")
	}
	return "", err
}
