package binder

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_JSON_Binding_Bind(t *testing.T) {
	t.Parallel()

	b := &jsonBinding{}
	require.Equal(t, "json", b.Name())

	type Post struct {
		Title string `json:"title"`
	}

	type User struct {
		Name  string `json:"name"`
		Posts []Post `json:"posts"`
		Age   int    `json:"age"`
	}
	var user User

	err := b.Bind(json.NewDecoder(bytes.NewReader([]byte(`{"name":"john","age":42,"posts":[{"title":"post1"},{"title":"post2"},{"title":"post3"}]}`))), &user)
	require.NoError(t, err)
	require.Equal(t, "john", user.Name)
	require.Equal(t, 42, user.Age)
	require.Len(t, user.Posts, 3)
	require.Equal(t, "post1", user.Posts[0].Title)
	require.Equal(t, "post2", user.Posts[1].Title)
	require.Equal(t, "post3", user.Posts[2].Title)
}

func Benchmark_JSON_Binding_Bind(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	binder := &jsonBinding{}

	type User struct {
		Name  string   `json:"name"`
		Posts []string `json:"posts"`
		Age   int      `json:"age"`
	}

	var user User
	var err error
	for i := 0; i < b.N; i++ {
		err = binder.Bind(json.NewDecoder(bytes.NewReader([]byte(`{"name":"john","age":42,"posts":["post1","post2","post3"]}`))), &user)
	}

	require.NoError(b, err)
	require.Equal(b, "john", user.Name)
	require.Equal(b, 42, user.Age)
	require.Len(b, user.Posts, 3)
	require.Equal(b, "post1", user.Posts[0])
	require.Equal(b, "post2", user.Posts[1])
	require.Equal(b, "post3", user.Posts[2])
}
