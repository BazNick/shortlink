package storage

type Storage interface {
	AddHash(hash, link string) (string, error)
	GetHash(hash string) string
	CheckValExists(link string) bool
}