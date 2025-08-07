package main

import (
	"article-versioning-api/core/usecase"
	"article-versioning-api/handler"
	userrepository "article-versioning-api/repository/user"
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()

	dbDSN := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		panic(err)
	}
	log.Printf("Connected to database: %v", dbDSN)
	defer db.Close()

	userRepo := userrepository.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase()
	userUsecase := usecase.NewUserUsecase(userRepo, authUsecase)

	userHandler := handler.NewUserHandler(userUsecase)
	authHandler := handler.NewAuthHandler(authUsecase)
	articleHandler := handler.NewArticleHandler()

	authenticatedRoute := router.Group("/")
	authenticatedRoute.Use(authHandler.VerifyToken)
	{
		authenticatedRoute.POST("/article", articleHandler.CreateArticle)
	}

	router.POST("/user/register", userHandler.RegisterUser)
	router.POST("/user/login", userHandler.Login)

	router.Run()
}
