package storage

type Storage interface {
	AddHash(hash, link, userID string) (string, error)
	GetHash(hash string) string
	CheckValExists(link string) bool
}