package storage

import "io"

type NoneStorage struct{}

func NewNoneStorage() *NoneStorage {
	return &NoneStorage{}
}

func (s *NoneStorage) Read() (map[string]string, error) {
	return make(map[string]string), nil
}

func (s *NoneStorage) Get(key string) (string, bool) {
	return "", false
}

func (s *NoneStorage) Put(key string, target string) {
}

func (s *NoneStorage) Delete(key string) {
}

func (s *NoneStorage) Update(key string, value string) {
}

func (s *NoneStorage) ReplaceConfig(reader io.Reader) (map[string]string, error) {
	return parseLinksFile(reader)
}

func (s *NoneStorage) GetReloadChannel() <-chan bool {
	return nil
}
