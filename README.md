# Chix

This package provides some methods that Chi lacks, such as binding and rendering.

A lot of the code in this package comes from [Fiber](https://github.com/gofiber/fiber), the last synchronized version: [bc4c920ea6b36d2b9d0396853a640b8b043951b5](https://github.com/gofiber/fiber/commit/bc4c920ea6b36d2b9d0396853a640b8b043951b5).

## Guides

### Custom Encoders and Decoders

Chix supports custom JSON/XML encoders and decoders. Here's an example:

```go
import (
    "encoding/json"
    "encoding/xml"

    "github.com/go-rat/chix"
)

func init() {
    chix.JSONEncoder = json.NewEncoder
    chix.JSONDecoder = json.NewDecoder
    chix.XMLEncoder = xml.NewEncoder
    chix.XMLDecoder = xml.NewDecoder
}
```

### Binding

#### Support Binders

- [Form](binder/form.go)
- [Query](binder/query.go)
- [URI](binder/uri.go)
- [Header](binder/header.go)
- [Cookie](binder/cookie.go)
- [JSON](binder/json.go)
- [XML](binder/xml.go)

#### Binding into a Struct

Chix supports binding request data directly into a struct using [gofiber/schema](https://github.com/gofiber/schema). Here's an example:

```go
// Field names must start with an uppercase letter
type Person struct {
	Name string `json:"name" xml:"name" form:"name"`
	Pass string `json:"pass" xml:"pass" form:"pass"`
}

router.Post("/", func(w http.ResponseWriter, r *http.Request) {
	p := new(Person)
	bind := chix.NewBind(r)
	defer bind.Release()

	if err := bind.Body(p); err != nil {
		return err
	}

	log.Println(p.Name) // Output: john
	log.Println(p.Pass) // Output: doe

	// Additional logic...
})

// Run tests with the following curl commands:

// JSON
curl -X POST -H "Content-Type: application/json" --data "{\"name\":\"john\",\"pass\":\"doe\"}" localhost:3000

// XML
curl -X POST -H "Content-Type: application/xml" --data "<login><name>john</name><pass>doe</pass></login>" localhost:3000

// URL-Encoded Form
curl -X POST -H "Content-Type: application/x-www-form-urlencoded" --data "name=john&pass=doe" localhost:3000

// Multipart Form
curl -X POST -F name=john -F pass=doe http://localhost:3000

// Query Parameters
curl -X POST "http://localhost:3000/?name=john&pass=doe"
```

#### Binding into q Map

Chix allows binding request data into a `map[string]string` or `map[string][]string`. Here's an example:

```go
router.Get("/", func(w http.ResponseWriter, r *http.Request) {
	params := make(map[string][]string)
	bind := chix.NewBind(r)
	defer bind.Release()

	if err := bind.Query(params); err != nil {
		return err
	}

	log.Println(params["name"])     // Output: [john]
	log.Println(params["pass"])     // Output: [doe]
	log.Println(params["products"]) // Output: [shoe hat]

	// Additional logic...
})

// Run tests with the following curl command:

curl "http://localhost:3000/?name=john&pass=doe&products=shoe&products=hat"
```

### Render

#### Support Methods

- ContentType
- Status
- Header
- Cookie
- WithoutCookie
- Redirect
- RedirectPermanent
- PlainText
- Data
- HTML
- JSON
- JSONP
- XML
- NoContent
- Stream
- EventStream
- SSEvent
- File
- Download
- Flush
- Hijack
- Release

#### Render a JSON

```go
router.Get("/", func(w http.ResponseWriter, r *http.Request) {
	render := chix.NewRender(w)
	defer render.Release()
	render.JSON(chix.M{
		"hello": "world",
	})
})
```
