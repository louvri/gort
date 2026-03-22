# gort

`gort` is a Go library that provides common server transport utilities — authentication middleware, JWT helpers, timezone extraction, and maintenance mode support.

Each submodule is independently importable so you only pull in the dependencies you need.

## Modules

| Module | Install | Description |
|--------|---------|-------------|
| [common](./common) | `go get github.com/louvri/gort/common` | Framework-agnostic JWT parsing, bearer token extraction, and timezone utilities |
| [echo](./echo) | `go get github.com/louvri/gort/echo` | Authentication and maintenance middleware for [Echo](https://github.com/labstack/echo) |
| [gin](./gin) | `go get github.com/louvri/gort/gin` | Authentication and maintenance middleware for [Gin](https://github.com/gin-gonic/gin) |

## Quick Start

### Gin

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    gortGin "github.com/louvri/gort/gin"
)

func main() {
    r := gin.Default()

    // JWT authentication
    r.Use(gortGin.JWTAuthValidatorMiddleware("your-secret", "Unauthorized", true, true))

    r.GET("/hello", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "hello world"})
    })

    r.Run()
}
```

### Echo

```go
package main

import (
    "net/http"

    "github.com/labstack/echo/v4"
    gortEcho "github.com/louvri/gort/echo"
)

func main() {
    e := echo.New()

    // JWT authentication
    e.Use(gortEcho.JWTAuthValidatorMiddleware("your-secret", "Unauthorized", true, true))

    e.GET("/hello", func(c echo.Context) error {
        return c.String(http.StatusOK, "hello world")
    })

    e.Logger.Fatal(e.Start(":8080"))
}
```

See each submodule's README for detailed API documentation.

## License

gort is released under the [MIT License](http://www.opensource.org/licenses/mit-license.php).
