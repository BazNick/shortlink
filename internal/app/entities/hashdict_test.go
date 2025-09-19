package entities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckValExists(t *testing.T) {
	tests := []struct {
		name string
		hd   HashDict
		link string
		want bool
	}{
		{
			name: "Value exists in the hash dictionary",
			hd: HashDict{
				Dict: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				RevDict: map[string]string{
					"value1": "key1",
					"value2": "key2",
				},
				Users: map[string]bool{},
			},
			link: "value2",
			want: true,
		},
		{
			name: "Value does not exist in the hash dictionary",
			hd: HashDict{
				Dict: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				RevDict: map[string]string{
					"value1": "key1",
					"value2": "key2",
				},
				Users: map[string]bool{},
			},
			link: "nonexistent_value",
			want: false,
		},
		{
			name: "Empty hash dictionary",
			hd: HashDict{
				Dict:    map[string]string{},
				RevDict: map[string]string{},
				Users:   map[string]bool{},
			},
			link: "any_value",
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.hd.CheckValExists(test.link)
			require.Equal(t, test.want, got)
		})
	}
}

func TestHashDict_AddHash(t *testing.T) {
	tests := []struct {
		name   string
		h      HashDict
		hash   string
		link   string
		userID string
	}{
		{
			name: "Adding a new key-value pair to an empty hash dictionary",
			h: HashDict{
				Dict:    map[string]string{},
				RevDict: map[string]string{},
				Users:   map[string]bool{},
			},
			hash:   "new_key",
			link:   "new_value",
			userID: "test",
		},
		{
			name: "Adding a new key-value pair to a non-empty hash dictionary",
			h: HashDict{
				Dict: map[string]string{
					"existing_key": "existing_value",
				},
				RevDict: map[string]string{
					"existing_value": "existing_key",
				},
				Users: map[string]bool{},
			},
			hash:   "another_new_key",
			link:   "another_new_value",
			userID: "test",
		},
		{
			name: "Overwriting existing value with a new one",
			h: HashDict{
				Dict: map[string]string{
					"existing_key": "old_value",
				},
				RevDict: map[string]string{
					"old_value": "existing_key",
				},
				Users: map[string]bool{},
			},
			hash:   "existing_key",
			link:   "updated_value",
			userID: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.AddHash(tt.hash, tt.link, tt.userID)
			require.Contains(t, tt.h.Dict, tt.hash)
			require.Equal(t, tt.link, tt.h.Dict[tt.hash])
		})
	}
}

func TestHashDict_GetHash(t *testing.T) {
	tests := []struct {
		name string
		h    HashDict
		hash string
		want string
	}{
		{
			name: "Getting value for existing key",
			h: HashDict{
				Dict: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				RevDict: map[string]string{
					"value1": "key1",
					"value2": "key2",
				},
				Users: map[string]bool{},
			},
			hash: "key1",
			want: "value1",
		},
		{
			name: "Getting value for non-existing key",
			h: HashDict{
				Dict: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				RevDict: map[string]string{
					"value1": "key1",
					"value2": "key2",
				},
				Users: map[string]bool{},
			},
			hash: "non_existing_key",
			want: "",
		},
		{
			name: "Getting value from an empty hash dictionary",
			h: HashDict{
				Dict:    map[string]string{},
				RevDict: map[string]string{},
				Users:   map[string]bool{},
			},
			hash: "any_key",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.h.GetHash(tt.hash)
			require.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkHashDict_AddHash(b *testing.B) {
	hd := NewHashDict()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hd.AddHash("hash"+string(rune(i)), "link"+string(rune(i)), "user"+string(rune(i)))
	}
}

func BenchmarkHashDict_GetHash(b *testing.B) {
	hd := NewHashDict()

	for i := 0; i < 1000; i++ {
		hd.AddHash("hash"+string(rune(i)), "link"+string(rune(i)), "user"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hd.GetHash("hash" + string(rune(i%1000)))
	}
}

func BenchmarkHashDict_CheckValExists(b *testing.B) {
	hd := NewHashDict()

	for i := 0; i < 1000; i++ {
		hd.AddHash("hash"+string(rune(i)), "link"+string(rune(i)), "user"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hd.CheckValExists("link" + string(rune(i%1000)))
	}
}
