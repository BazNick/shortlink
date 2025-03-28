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
			},
			link: "nonexistent_value",
			want: false,
		},
		{
			name: "Empty hash dictionary",
			hd: HashDict{
				Dict: map[string]string{},
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
		name string
		h    HashDict
		hash string
		link string
	}{
		{
			name: "Adding a new key-value pair to an empty hash dictionary",
			h: HashDict{
				Dict: map[string]string{},
			},
			hash: "new_key",
			link: "new_value",
		},
		{
			name: "Adding a new key-value pair to a non-empty hash dictionary",
			h: HashDict{
				Dict: map[string]string{
					"existing_key": "existing_value",
				},
			},
			hash: "another_new_key",
			link: "another_new_value",
		},
		{
			name: "Overwriting existing value with a new one",
			h: HashDict{
				Dict: map[string]string{
					"existing_key": "old_value",
				},
			},
			hash: "existing_key",
			link: "updated_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.AddHash(tt.hash, tt.link)
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
			},
			hash: "non_existing_key",
			want: "",
		},
		{
			name: "Getting value from an empty hash dictionary",
			h: HashDict{
				Dict: map[string]string{},
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
