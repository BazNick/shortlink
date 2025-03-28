package entities

type HashDict struct {
	Dict map[string]string
}

func NewHashDict() *HashDict {
	return &HashDict{
		Dict: make(map[string]string),
	}
}

func (hasdDict *HashDict) AddHash(hash, link string) {
	hasdDict.Dict[hash] = link
}

func (hasdDict *HashDict) GetHash(hash string) string {
	if val, ok := hasdDict.Dict[hash]; ok {
		return val
	}
	return ""
}

func (hasdDict *HashDict) CheckValExists(link string) bool {
	for _, v := range hasdDict.Dict {
		if v == link {
			return true
		}
	}
	return false
}
