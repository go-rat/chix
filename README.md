# Chi Binding

This is a fork and modified version of Fiber for Chi router.

Current Version [260c5e5406874e6d9a48ec2ef2c862d64a530e0b](https://github.com/gofiber/fiber/commit/260c5e5406874e6d9a48ec2ef2c862d64a530e0b)

## Support Binders

- [Form](form.go)
- [Query](query.go)
- [URI](uri.go)
- [Header](header.go)
- [Cookie](cookie.go)
- [JSON](json.go)
- [XML](xml.go)

## Guides

### Binding into the Struct

Fiber supports binding into the struct with [gorilla/schema](https://github.com/gorilla/schema). Here's an example:

```go
// Field names should start with an uppercase letter
type Person struct {
    Name string `json:"name" xml:"name" form:"name"`
    Pass string `json:"pass" xml:"pass" form:"pass"`
}

app.Post("/", func(w http.ResponseWriter, r *http.Request) {
    p := new(Person)

    if err := binding.New(r).Body(p); err != nil {
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

### Binding into the Map

Fiber supports binding into the `map[string]string` or `map[string][]string`. Here's an example:

```go
app.Get("/", func(w http.ResponseWriter, r *http.Request) {
    p := make(map[string][]string)

    if err := binding.New(r).Query(p); err != nil {
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

### Behaviors of Should/Must

Normally, Fiber returns binder error directly. However; if you want to handle it automatically, you can prefer `Must()`.

If there's an error it'll return error and 400 as HTTP status. Here's an example for it:

```go
// Field names should start with an uppercase letter
type Person struct {
    Name string `json:"name,required"`
    Pass string `json:"pass"`
}

app.Get("/", func(w http.ResponseWriter, r *http.Request) {
    p := new(Person)

    if err := binding.New(r).Must().JSON(p); err != nil {
        return err 
        // Status code: 400 
        // Response: Bad request: name is empty
    }

    // ...
})

// Run tests with the following curl command:

// curl -X GET -H "Content-Type: application/json" --data "{\"pass\":\"doe\"}" localhost:3000
```
