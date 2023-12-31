package main

import (
	"buddle-server/handler"
	"buddle-server/internal/app/api"
	"buddle-server/internal/db"
	"buddle-server/internal/log"
	"buddle-server/internal/s3"
	"buddle-server/middleware"
	"buddle-server/repository"
	"buddle-server/service"
	"context"
	"github.com/labstack/echo/v4"
	md "github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configPath = os.Getenv("BD_CONFIG")
)

type server struct {
	echo *echo.Echo
	db   *gorm.DB

	// Handlers
	productHandler      handler.ProductHandler
	userHandler         handler.UserHandler
	afterServiceHandler handler.AfterServiceHandler

	// Services
	productService service.ProductService
	userService    service.UserService
	afterService   service.AfterService

	// Repositories
	repo repository.Repository

	// S3
	fileBucket *s3.S3
}

func NewServer() (*server, error) {
	s := &server{echo: echo.New()}

	if err := s.initConfig(); err != nil {
		return nil, errors.Wrap(err, "failed to init config")
	}
	if err := s.initRepositories(); err != nil {
		return nil, errors.Wrap(err, "failed to init repository")
	}
	if err := s.initServices(); err != nil {
		return nil, errors.Wrap(err, "failed to init service")
	}
	if err := s.initHandlers(); err != nil {
		return nil, errors.Wrap(err, "failed to init handlers")
	}

	s.initRoutes()

	return s, nil
}

func (s *server) initHandlers() (err error) {
	if s.productHandler, err = handler.NewProductHandler(s.productService); err != nil {
		return errors.Wrap(err, "failed init product handler")
	}
	if s.afterServiceHandler, err = handler.NewAfterHandler(s.afterService); err != nil {
		return errors.Wrap(err, "failed init product handler")
	}
	if s.userHandler, err = handler.NewUserHandler(s.userService); err != nil {
		return errors.Wrap(err, "failed init login handler")
	}
	return
}

func (s *server) initServices() (err error) {
	if s.productService, err = service.NewProductService(s.repo, s.fileBucket); err != nil {
		return errors.Wrap(err, "failed init product services")
	}
	if s.afterService, err = service.NewAfterService(s.repo, s.fileBucket); err != nil {
		return errors.Wrap(err, "failed init product services")
	}
	if s.userService, err = service.NewUserService(s.repo); err != nil {
		return errors.Wrap(err, "failed init user services")
	}
	return
}

func (s *server) initRepositories() (err error) {
	s.repo, err = repository.NewRepository()
	if err != nil {
		return errors.Wrap(err, "failed to create repository")
	}
	return
}

func (s *server) initRoutes() {
	s.echo.Use(md.CORS())

	s.echo.GET("/healthcheck", func(echoCtx echo.Context) error {
		return echoCtx.String(http.StatusOK, "OK")
	})

	v1 := s.echo.Group(
		"/v1",
		middleware.CustomContext,
		middleware.WithDB("", s.db),
	)

	jwtMiddleWare := md.JWTWithConfig(md.JWTConfig{
		SigningKey:  []byte(api.Config().Jwt.SecretKey),
		TokenLookup: "header:access-token,query:access-token",
	})

	v1Product := v1.Group("/product", jwtMiddleWare)
	{
		v1Product.POST("", s.productHandler.CreateProduct)
		v1Product.GET("/manage", s.productHandler.FindProductList)
		v1Product.GET("/receipt", s.productHandler.DownloadReceipt)
	}

	v1ProductRegist := v1.Group("/product-regist")
	{
		v1ProductRegist.GET("", s.productHandler.GetAuthProduct)
		v1ProductRegist.POST("", s.productHandler.AuthProduct)
		v1ProductRegist.PUT("/:product_regist_seq", s.productHandler.ModAuthProduct)
		v1ProductRegist.POST("/:product_regist_seq/cancel", s.productHandler.CancelAuthProduct)
	}

	v1AfterService := v1.Group("/as")
	{
		v1AfterService.POST("", s.afterServiceHandler.Create)
		v1AfterService.GET("", s.afterServiceHandler.FindAfterServiceInfo)
		v1AfterService.GET("/manage", s.afterServiceHandler.FindAfterServiceManagerInfo)
		v1AfterService.GET("/file", s.afterServiceHandler.DownloadFile)
	}

	v1user := v1.Group("/user")
	{
		v1user.POST("", s.userHandler.SignUp)
		v1user.POST("/login", s.userHandler.SignIn)
	}
}

func (s *server) initConfig() (err error) {
	if err = api.InitConfig(configPath); err != nil {
		return errors.Wrapf(err, "Load config file path = %s", configPath)
	}

	conf := api.Config()
	if err = log.Init(conf.Logger); err != nil {
		return errors.Wrap(err, "Init logger")
	}

	if s.db, err = db.Connect(conf.DB); err != nil {
		return errors.Wrap(err, "Init db")
	}

	if s.fileBucket, err = s3.New(conf.FileBucket); err != nil {
		return errors.Wrap(err, "Init file bucket")
	}

	return nil
}

func (s *server) start() error {
	go func() {
		if err := s.echo.Start(":1202"); err != nil {
			logrus.Errorf("shutting down the Buddle go api server [ err:%+v ]", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		os.Interrupt,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
		syscall.SIGTERM,
	)
	sig := <-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.echo.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown the Directcloud go api server [ sig : %+v, err : %+v ]", sig, err)
	}

	return nil
}
