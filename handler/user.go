package handler

import (
	"buddle-server/internal/app/api"
	"buddle-server/internal/jwt"
	"buddle-server/middleware"
	"buddle-server/model"
	"buddle-server/service"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
)

type UserHandler interface {
	SignUp(c echo.Context) error // 회원가입
	SignIn(c echo.Context) error // 로그인
}

type userHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) (UserHandler, error) {
	if userService == nil {
		return nil, errors.New("user service is nil")
	}

	return &userHandler{
		userService: userService,
	}, nil
}

func (l userHandler) SignUp(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	user := new(model.User)

	if err := ctx.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "bad request",
		})
	}

	if err := user.SignUpCheck(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "bad request",
		})
	}

	if err := l.userService.SignUp(ctx.GoContext(), user); err != nil {
		return errors.Wrap(err, "failed to sign up user")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "success",
	})
}

func (l userHandler) SignIn(c echo.Context) error {
	ctx, err := middleware.UpgradeContext(c)
	if err != nil {
		return errors.Wrap(err, "upgrade context")
	}

	user := new(model.User)

	if err := ctx.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "bad request",
		})
	}

	resp, err := l.userService.SignIn(ctx.GoContext(), user)
	if err != nil {
		return errors.Wrap(err, "failed to sign up user")
	}

	// 토큰 발행
	accessToken, err := jwt.CreateJWT(user.Id, api.Config().Jwt.SecretKey)
	if err != nil {
		return echo.ErrInternalServerError
	}

	resp.Data = struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: accessToken,
	}

	return c.JSON(http.StatusOK, resp)
}
