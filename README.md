# gort
<hr style="border:1px solid #444; margin-top: -0.5em;">  

`gort` is a library that provides simple and common golang's server transport.   
Currently supports for [ECHO](https://github.com/labstack/echo) and [GIN](https://github.com/gin-gonic/gin).  
  
### Installation
<hr style="border:1px solid #444; margin-top: -0.5em;">  

Get the code with:  
```
$ go get github.com/louvri/gort
```
Then use it on your codes:
```
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    commonGin "github.com/louvri/gort/gin"
)

func main() {
    r := gin.Default()
    r.Use(commonGin.JWTAuthValidatorMiddleware("testing", "Invalid/expired token", true))
    r.GET("/hello-world", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "hello world",
        })
    })
    r.Run()
}
```
  
### Usage
<hr style="border:1px solid #444; margin-top: -0.5em;">  

 - Common
   ```
    package main

    import (
        "net/http"
        commonTransport "github.com/louvri/gort/common"
        ...
    )
    
    func main() {
        ...
        timeDifference := commonTransport.ExtractTimeZoneInSecondsFromHeader(c.Request())
        // do something with timeDifference
        ...
        authHeaderValue := commonTransport.GetAuthorizationHeaderValue(c.Request)
        // do something with authHeaderValue
        ...
        bearerToken := commonTransport.GetBearerToken(c.Request)
        // do something with bearerToken
        ...
        claims := commonTransport.GetMapClaimsFromJWT("testing", bearerToken, true)
        // do something with claims
        ...
    }
   ```
 - with Gin
   ```
    package main

    import (
        "net/http"
        "github.com/gin-gonic/gin"
        commonGin "github.com/louvri/gort/gin"
    )
    
    func main() {
        r := gin.Default()
        publicGroup := router.Group("/public")
        publicGroup.Use(commonGin.JWTAuthValidatorMiddleware("testing", "Invalid/expired token", true))
        publicGroup.GET("/hello-world", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "message": "hello world",
            })
        })
        internalGroup := router.Group("/internal")
        internalGroup.Use(commonGin.ServerKeyAuthValidatorMiddleware("Authorization-Id", "zJ11tGd1Wr1v", "hU9F8b2CdvrF", "Invalid token"))
        internalGroup.GET("/hello-world", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "message": "hello world",
            })
        })
        r.Run()
    }
   ```
 - with Echo
   ```
    package main

    import (
        "net/http"
        "github.com/labstack/echo/v4"
        "github.com/labstack/echo/v4/middleware"
        commonEcho "github.com/louvri/gort/echo"
    )
    
    func main() {
        e := echo.New()
        e.Use(middleware.Recover())
        publicGroup := e.Group("/public")
        publicGroup.Use(commonEcho.JWTAuthValidatorMiddleware("testing", "Invalid/expired token", true))
        publicGroup.GET("/hello-world", func(c echo.Context) error {
            return c.String(http.StatusOK, "Hello World")
        })
        internalGroup := e.Group("/internal")
        internalGroup.Use(commonEcho.ServerKeyAuthValidatorMiddleware("Authorization-Id", "zJ11tGd1Wr1v", "hU9F8b2CdvrF", "Invalid token"))
        internalGroup.GET("/hello-world", func(c echo.Context) error {
            return c.String(http.StatusOK, "Hello World")
        })
        e.Logger.Fatal(e.Start(":1323"))
    }
   ```
  

### License
<hr style="border:1px solid #444; margin-top: -0.5em;">  

gort is released under the [MIT License](http://www.opensource.org/licenses/mit-license.php)
