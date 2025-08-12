package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"

	_ "github.com/lib/pq"
)

func main() {
	// cron job update tag trending score value (using exponential decay)
	cronJobSchedule := os.Getenv("UPDATE_TAG_TRENDING_SCORE_SCHEDULE")

	c := cron.New()
	_, err := c.AddFunc(cronJobSchedule, func() {
		err := FetchAndStoreData()
		if err != nil {
			log.Printf("error cron job update tag trending score: %v", err.Error())
		}
	})
	if err != nil {
		log.Fatalf("error cron job update tag trending score: %v", err.Error())
	}

	c.Start()
	log.Println("[info] start cron job update tag trending score")

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("[info] shutting down cron job update tag trending score")
	c.Stop()
	log.Println("[info] cron job update tag trending score stopped")
}

func FetchAndStoreData() error {
	clientURL := "http://app:8080/tags/trending-score"

	req, err := http.NewRequest(http.MethodPut, clientURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error with status code: %v", response.StatusCode)
	}

	log.Printf("[info] cron job update tag trending score is successful")
	return nil
}
