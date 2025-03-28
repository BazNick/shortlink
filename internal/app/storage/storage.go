package storage

type Storage interface {
	AddHash(hash, link string)
	GetHash(hash string) string
	CheckValExists(link string) bool
}