package entities

type HashDict map[string]string

var Hash = make(HashDict, 1)

func (h HashDict) AddHash(hash, link string) {
	h[hash] = link
}

func (h HashDict) GetHash(hash string) string {
	if val, ok := h[hash]; ok {
		return val
	}
	return ""
}

func CheckValExists(hd HashDict, link string) bool {
	for _, v := range hd {
		if v == link {
			return true
		}
	}
	return false
}