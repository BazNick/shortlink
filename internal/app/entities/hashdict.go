package entities

type HashDict struct {
	Dict    map[string]string
	RevDict map[string]string
}

func NewHashDict() *HashDict {
	return &HashDict{
		Dict:    make(map[string]string),
		RevDict: make(map[string]string),
	}
}

func (hasdDict *HashDict) AddHash(hash, link, userID string) (string, error) {
	hasdDict.Dict[hash] = link
	hasdDict.RevDict[link] = hash
	return "", nil
}

func (hasdDict *HashDict) GetHash(hash string) string {
	return hasdDict.Dict[hash]
}

func (hasdDict *HashDict) CheckValExists(link string) bool {
	_, exists := hasdDict.RevDict[link]
	return exists
}
