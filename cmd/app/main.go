package main

import (
	"article-versioning-api/config"
	"article-versioning-api/core/usecase"
	"article-versioning-api/handler"
	articlerepository "article-versioning-api/repository/article"
	tagrepository "article-versioning-api/repository/tag"
	userrepository "article-versioning-api/repository/user"
	transactionutil "article-versioning-api/utils/transaction"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()

	cfg := config.GetConfig()

	dbDSN := cfg.DatabaseUrl
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		panic(err)
	}
	log.Printf("Connected to database: %v", dbDSN)
	defer db.Close()

	// TODO
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	transactionPkg := transactionutil.NewConnection(gormDB)

	userRepo := userrepository.NewUserRepository(db, cfg)
	articleRepo := articlerepository.NewArticleRepository(db, cfg, gormDB)
	tagRepo := tagrepository.NewTagRepository(db, gormDB)

	authUsecase := usecase.NewAuthUsecase()
	userUsecase := usecase.NewUserUsecase(userRepo, authUsecase)
	articleUsecase := usecase.NewArticleUsecase(articleRepo, tagRepo, transactionPkg, cfg)
	tagUsecase := usecase.NewTagUsecase(tagRepo, transactionPkg, cfg)

	userHandler := handler.NewUserHandler(userUsecase)
	authHandler := handler.NewAuthHandler(authUsecase)
	articleHandler := handler.NewArticleHandler(articleUsecase)
	tagHandler := handler.NewTagHandler(tagUsecase)

	writerRoute := router.Group("/")
	writerRoute.Use(authHandler.VerifyToken)
	writerRoute.Use(authHandler.VerifyRole([]string{"writer"}))
	{
		writerRoute.POST("/articles", articleHandler.CreateArticle)
		writerRoute.POST("/articles/:serial/version", articleHandler.CreateArticleVersion)
	}

	adminWriterRoute := router.Group("/")
	adminWriterRoute.Use(authHandler.VerifyToken)
	adminWriterRoute.Use(authHandler.VerifyRole([]string{"admin", "writer"}))
	{
		adminWriterRoute.PATCH("articles/:serial/versions/:versionSerial/status", articleHandler.UpdateArticleVersionStatus)
		adminWriterRoute.DELETE("articles/:serial", articleHandler.DeleteArticle)
		adminWriterRoute.GET("/articles/:serial/latest-details", articleHandler.GetArticleLatestDetail)
		adminWriterRoute.GET("/articles/:serial/versions", articleHandler.GetVersionsByArticleSerial)
		adminWriterRoute.GET("/articles/versions/:versionSerial", articleHandler.GetVersionBySerial)

		adminWriterRoute.POST("/tags", tagHandler.CreateTag)
		adminWriterRoute.GET("/tags", tagHandler.GetTags)
		adminWriterRoute.GET("/tags/:serial", tagHandler.GetTagBySerial)
	}

	authenticatedRoute := router.Group("/")
	authenticatedRoute.Use(authHandler.VerifyToken)
	{
		authenticatedRoute.GET("/articles", articleHandler.GetArticles)
	}

	router.POST("/users/register", userHandler.RegisterUser)
	router.POST("/users/login", userHandler.Login)

	router.PUT("/tags/trending-score", articleHandler.UpdateTrendingScoreTags)

	router.Run()
}
