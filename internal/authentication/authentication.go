package authentication

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	ginserver "github.com/go-oauth2/gin-server"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/model"
	"gopkg.in/oauth2.v3"

	"github.com/gin-gonic/gin"
)

const (
	headerName = "X-Gotify-Key"
)

type (
	// AuthenticationValidator defines a type for authenticating a user
	AuthenticationValidator func() gin.HandlerFunc
	// UserSetter defines a type for setting a user object
	UserSetter func() gin.HandlerFunc
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
	GetUserByID(id uint) (*model.User, error)
}

// Authenticator is the provider for authentication middleware.
type Authenticator struct {
	DB                      Database
	Config                  configuration.Authentication
	AuthenticationValidator AuthenticationValidator
	UserSetter              UserSetter
}

type hasUserProperty func(user *model.User) bool

func (a *Authenticator) userFromBasicAuth(ctx *gin.Context) (*model.User, error) {
	if name, password, ok := ctx.Request.BasicAuth(); ok {
		if user, err := a.DB.GetUserByName(name); err != nil {
			return nil, err
		} else if user != nil && credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return user, nil
		} else {
			return nil, errors.New("credentials were invalid")
		}
	}

	return nil, errors.New("no credentials were supplied 1")
}

func (a *Authenticator) userFromToken(ctx *gin.Context) (*model.User, error) {
	ti, exists := ctx.Get(ginserver.DefaultConfig.TokenKey)
	if !exists {
		return nil, errors.New("No token available")
	}

	token, ok := ti.(oauth2.TokenInfo)
	if !ok {
		return nil, errors.New("Wrong token format")
	}

	userID, err := strconv.ParseUint(token.GetUserID(), 10, 64)
	if err != nil {
		return nil, errors.New("User information of wrong format")
	}

	user, err := a.DB.GetUserByID(uint(userID))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *Authenticator) requireUserProperty(has hasUserProperty) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := errors.New("User not found")

		u, exists := ctx.Get("user")

		if !exists {
			log.Println("No user object in context")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		user, ok := u.(*model.User)

		if !ok {
			log.Println("User object from context has wrong format")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		if !has(user) {
			ctx.AbortWithError(http.StatusForbidden, errors.New("authentication failed"))
			return
		}
	}
}

// RequireUser returns a Gin middleware which requires valid user credentials to be supplied with the request.
func (a *Authenticator) RequireUser() gin.HandlerFunc {
	return a.UserSetter()
}

// RequireAdmin returns a Gin middleware which requires valid admin credentials to be supplied with the request.
func (a *Authenticator) RequireAdmin() gin.HandlerFunc {
	return a.requireUserProperty(func(user *model.User) bool {
		return user.IsAdmin
	})
}

func (a *Authenticator) tokenFromQueryOrHeader(ctx *gin.Context) string {
	if token := a.tokenFromQuery(ctx); token != "" {
		return token
	} else if token := a.tokenFromHeader(ctx); token != "" {
		return token
	}

	return ""
}

func (a *Authenticator) tokenFromQuery(ctx *gin.Context) string {
	return ctx.Request.URL.Query().Get("token")
}

func (a *Authenticator) tokenFromHeader(ctx *gin.Context) string {
	return ctx.Request.Header.Get(headerName)
}

// RequireApplicationToken returns a Gin middleware which requires an application token to be supplied with the request.
func (a *Authenticator) RequireApplicationToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := a.tokenFromQueryOrHeader(ctx)

		app, err := a.DB.GetApplicationByToken(token)
		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		ctx.Set("app", app)
	}
}

// RequireValidAuthentication returns a Gin middleware which requires a valid authentication
func (a *Authenticator) RequireValidAuthentication() gin.HandlerFunc {
	return a.AuthenticationValidator()
}

// SetAuthenticationValidator sets a function for handling authentication
func (a *Authenticator) SetAuthenticationValidator(f AuthenticationValidator) {
	a.AuthenticationValidator = f
}

// SetUserSetter sets a function that sets the user object in gin context
func (a *Authenticator) SetUserSetter(f UserSetter) {
	a.UserSetter = f
}
