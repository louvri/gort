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

Merging to `main` tags a release, `<module>/vX.Y.Z`, of every module the merge changed. Each change to the module since its last tag asks for a level, and the release takes the highest:

- A `Release-As: major|minor|patch` trailer on its own line sets the level of the change it is on (the key is case-insensitive). A change marked `Release-As: skip` never triggers a release by itself; it ships with the module's next change, at no lower a level than its own markers need. On a merge into `main`, the skip covers everything the merge brought in.
- Otherwise a breaking change (a `type!:` subject or a `BREAKING CHANGE:` footer) is a minor bump below v1.0.0 and a major one from then on; from v1.0.0 a `feat:` subject is a minor bump; everything else is a patch.
- A squash merge is one change: the `* subject` lines of its body count as subjects, and a `Release-As:` trailer anywhere in it applies to the whole pull request. Pull requests are squashed, so each one touching a module is a single change to it.
- A merge commit counts on its own subject and body, and one that changes the module itself (a hand-resolved conflict) asks for at least a patch.
- A module whose code is unchanged since its tag is not released, whatever its commits ask for. Once something else changes, a reverted change still counts: revert lines are not trusted, and a bump too large is the safe mistake.
- A module's first release is v0.1.0. From v2 on, its `go.mod` must declare the matching `/vN` module path, or the release is refused; if that major bump was not intended, tag the current `main` commit by hand with the version you want and later runs start from it (a tag on a commit that is not on `main` is ignored as a base, though its version number is still skipped past - so a `v1.x` tag pushed by hand off `main` moves a module past 1.0, after which a `feat:` is a minor bump and a breaking change a major one, refused until `go.mod` declares the `/v2` path).

The rules are pinned by `.github/scripts/next-version_test.sh` (needs git 2.38 or later).

## License

gort is released under the [MIT License](http://www.opensource.org/licenses/mit-license.php).
