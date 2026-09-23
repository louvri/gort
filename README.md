# gort

`gort` is a Go library that provides common server transport utilities — authentication middleware, JWT helpers, timezone extraction, and maintenance mode support.

Each submodule is independently importable so you only pull in the dependencies you need.

## Modules

| Module | Install | Go | Description |
|--------|---------|----|-------------|
| [common](./common) | `go get github.com/louvri/gort/common` | 1.25+ | Framework-agnostic JWT parsing, bearer token extraction, and timezone utilities |
| [echo](./echo) | `go get github.com/louvri/gort/echo` | 1.26+ | Authentication and maintenance middleware for [Echo](https://github.com/labstack/echo) |
| [gin](./gin) | `go get github.com/louvri/gort/gin` | 1.26+ | Authentication and maintenance middleware for [Gin](https://github.com/gin-gonic/gin) |

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

## Releasing

Merging to `main` tags a release, `<module>/vX.Y.Z`, of every module the merge changed. The level comes from the commits touching that module since its last tag:

- A `Release-As: major|minor|patch` trailer on its own line sets it explicitly; `Release-As: skip` publishes nothing, and those changes ship with the module's next change.
- Otherwise a breaking change (a `type!:` subject or a `BREAKING CHANGE:` footer) is a minor bump below v1.0.0 and a major one from then on; from v1.0.0 a `feat:` subject is a minor bump; everything else is a patch.
- Merges are squashed, so a pull request's trailers and markers apply to every module it touches.
- A module's first release is v0.1.0. From v2 on, its `go.mod` must declare the matching `/vN` module path, or the release is refused.

The rules are pinned by `.github/scripts/next-version_test.sh`.

## License

gort is released under the [MIT License](http://www.opensource.org/licenses/mit-license.php).
