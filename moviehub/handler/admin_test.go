package handler

import (
	"testing"

	"moviehub/store"
)

func TestValidateMovie(t *testing.T) {
	tests := []struct {
		name string
		in   store.MovieInput
		want string
	}{
		{
			name: "defaults and trims",
			in: store.MovieInput{
				Title: "  Example Movie  ",
			},
		},
		{
			name: "title required",
			in:   store.MovieInput{},
			want: "BAD_TITLE",
		},
		{
			name: "invalid date",
			in: store.MovieInput{
				Title:       "Example Movie",
				ReleaseDate: stringPointer("2025-02-30"),
			},
			want: "BAD_DATE",
		},
		{
			name: "invalid poster URL",
			in: store.MovieInput{
				Title:      "Example Movie",
				PosterPath: stringPointer("javascript:alert(1)"),
			},
			want: "BAD_POSTER",
		},
		{
			name: "poster URL accepted",
			in: store.MovieInput{
				Title:      "Example Movie",
				PosterPath: stringPointer("https://images.example/poster.jpg"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := validateMovie(&tt.in)
			if code != tt.want {
				t.Fatalf("validateMovie() code = %q, want %q", code, tt.want)
			}
		})
	}

	input := store.MovieInput{Title: "  Example Movie  "}
	if code, _ := validateMovie(&input); code != "" {
		t.Fatalf("valid movie rejected with %s", code)
	}
	if input.Title != "Example Movie" || input.OriginalTitle != "Example Movie" || input.OriginalLanguage != "en" {
		t.Fatalf("defaults not applied: %#v", input)
	}
}

func stringPointer(value string) *string {
	return &value
}
