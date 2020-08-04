package api

import (
	"errors"
	"net/http"

	"github.com/eikendev/pushbits/model"

	"github.com/gin-gonic/gin"
)

func getID(ctx *gin.Context) (uint, error) {
	id, ok := ctx.MustGet("id").(uint)
	if !ok {
		err := errors.New("an error occured while retrieving ID from context")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return 0, err
	}

	return id, nil
}

func getApplication(ctx *gin.Context, db Database) (*model.Application, error) {
	id, err := getID(ctx)
	if err != nil {
		return nil, err
	}

	application, err := db.GetApplicationByID(id)
	if success := successOrAbort(ctx, http.StatusNotFound, err); !success {
		return nil, err
	}

	return application, nil
}

func getUser(ctx *gin.Context, db Database) (*model.User, error) {
	id, err := getID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := db.GetUserByID(id)
	if success := successOrAbort(ctx, http.StatusNotFound, err); !success {
		return nil, err
	}

	return user, nil
}