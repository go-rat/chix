package binder

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_XMLBinding_Bind(t *testing.T) {
	t.Parallel()

	b := &xmlBinding{}
	require.Equal(t, "xml", b.Name())

	type Posts struct {
		XMLName xml.Name `xml:"post"`
		Title   string   `xml:"title"`
	}

	type User struct {
		Name   string  `xml:"name"`
		Ignore string  `xml:"-"`
		Posts  []Posts `xml:"posts>post"`
		Age    int     `xml:"age"`
	}

	user := new(User)
	err := b.Bind(xml.NewDecoder(bytes.NewReader([]byte(`
		<user>
			<name>john</name>
			<age>42</age>
			<ignore>ignore</ignore>
			<posts>
				<post>
					<title>post1</title>
				</post>
				<post>
					<title>post2</title>
				</post>
			</posts>
		</user>
	`))), user)
	require.NoError(t, err)
	require.Equal(t, "john", user.Name)
	require.Equal(t, 42, user.Age)
	require.Empty(t, user.Ignore)

	require.Len(t, user.Posts, 2)
	require.Equal(t, "post1", user.Posts[0].Title)
	require.Equal(t, "post2", user.Posts[1].Title)
}

func Test_XMLBinding_Bind_error(t *testing.T) {
	t.Parallel()
	b := &xmlBinding{}

	type User struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}

	user := new(User)
	err := b.Bind(xml.NewDecoder(bytes.NewReader([]byte(`
		<user>
			<name>john</name>
			<age>42</age>
			<unknown>unknown</unknown>
		</user
	`))), user)
	require.Error(t, err)
}

func Benchmark_XMLBinding_Bind(b *testing.B) {
	b.ReportAllocs()

	binder := &xmlBinding{}

	type Posts struct {
		XMLName xml.Name `xml:"post"`
		Title   string   `xml:"title"`
	}

	type User struct {
		Name  string  `xml:"name"`
		Posts []Posts `xml:"posts>post"`
		Age   int     `xml:"age"`
	}

	user := new(User)
	data := []byte(`
		<user>
			<name>john</name>
			<age>42</age>
			<ignore>ignore</ignore>
			<posts>
				<post>
					<title>post1</title>
				</post>
				<post>
					<title>post2</title>
				</post>
			</posts>
		</user>
	`)

	b.StartTimer()

	var err error
	for i := 0; i < b.N; i++ {
		err = binder.Bind(xml.NewDecoder(bytes.NewReader(data)), user)
	}
	require.NoError(b, err)

	user = new(User)
	err = binder.Bind(xml.NewDecoder(bytes.NewReader(data)), user)
	require.NoError(b, err)

	require.Equal(b, "john", user.Name)
	require.Equal(b, 42, user.Age)

	require.Len(b, user.Posts, 2)
	require.Equal(b, "post1", user.Posts[0].Title)
	require.Equal(b, "post2", user.Posts[1].Title)
}
