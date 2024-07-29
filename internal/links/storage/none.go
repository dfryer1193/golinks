package storage

type NoneStorage struct{}

func NewNoneStorage() *NoneStorage {
	return &NoneStorage{}
}

func (s *NoneStorage) Read() map[string]string {
	return nil
}

func (s *NoneStorage) Put(key string, target string) {
}

func (s *NoneStorage) Delete(key string) {
}

func (s *NoneStorage) Update(key string, value string) {
}

func (s *NoneStorage) GetReloadChannel() <-chan bool {
	return nil
}
