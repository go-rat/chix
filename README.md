# Chix

This package provides some methods that Chi lacks, such as binding and rendering, and it's a lightweight package that doesn't have any dependencies.

A lot of the code in this package comes from [Fiber](https://github.com/gofiber/fiber), the last synchronized version: [260c5e5406874e6d9a48ec2ef2c862d64a530e0b](https://github.com/gofiber/fiber/commit/260c5e5406874e6d9a48ec2ef2c862d64a530e0b).

## Guides

### Binding

#### Support Binders

- [Form](binder/form.go)
- [Query](binder/query.go)
- [URI](binder/uri.go)
- [Header](binder/header.go)
- [Cookie](binder/cookie.go)
- [JSON](binder/json.go)
- [XML](binder/xml.go)

#### Binding into the Struct

Chix supports binding into the struct with [gorilla/schema](https://github.com/gorilla/schema). Here's an example:

```go
// Field names should start with an uppercase letter
type Person struct {
    Name string `json:"name" xml:"name" form:"name"`
    Pass string `json:"pass" xml:"pass" form:"pass"`
}

router.Post("/", func(w http.ResponseWriter, r *http.Request) {
    p := new(Person)

    if err := chix.NewBind(r).Body(p); err != nil {
        return err
    }

    log.Println(p.Name) // john
    log.Println(p.Pass) // doe

    // ...
})

// Run tests with the following curl commands:

// curl -X POST -H "Content-Type: application/json" --data "{\"name\":\"john\",\"pass\":\"doe\"}" localhost:3000

// curl -X POST -H "Content-Type: application/xml" --data "<login><name>john</name><pass>doe</pass></login>" localhost:3000

// curl -X POST -H "Content-Type: application/x-www-form-urlencoded" --data "name=john&pass=doe" localhost:3000

// curl -X POST -F name=john -F pass=doe http://localhost:3000

// curl -X POST "http://localhost:3000/?name=john&pass=doe"
```

#### Binding into the Map

Chix supports binding into the `map[string]string` or `map[string][]string`. Here's an example:

```go
router.Get("/", func(w http.ResponseWriter, r *http.Request) {
    p := make(map[string][]string)

    if err := chix.NewBind(r).Query(p); err != nil {
        return err
    }

    log.Println(p["name"])     // john
    log.Println(p["pass"])     // doe
    log.Println(p["products"]) // [shoe, hat]

    // ...
})
// Run tests with the following curl command:

// curl "http://localhost:3000/?name=john&pass=doe&products=shoe,hat"
```

### Render

#### Support Methods

- Status
- Header
- Cookie
- WithoutCookie
- Redirect
- PlainText
- Data
- HTML
- JSON
- XML
- NoContent
- EventStream
- File
- Download
- Flush

#### Render a JSON

```go
router.Get("/", func(w http.ResponseWriter, r *http.Request) {
    return chix.NewRender(w).JSON(chix.M{
        "hello": "world",
    })
})
```
