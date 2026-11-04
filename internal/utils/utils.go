package utils

import (
	"fmt"
	"os"
)

func ValidatePathIsDir(path string) (string, error) {
	i, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !i.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}

	return path, nil
}

func ToSlice[K comparable, V any](m map[K]V) []V {
	s := make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}

func ToKeysSlice[K comparable, V any](m map[K]V) []K {
	s := make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
