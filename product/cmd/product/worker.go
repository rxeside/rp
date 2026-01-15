package main

import (
	"log"

	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	appservice "product/pkg/product/app/service"
	inframysql "product/pkg/product/infrastructure/mysql/repository"
	infratemporal "product/pkg/product/infrastructure/temporal"
)

func runWorker(config *config) *cli.Command {
	return &cli.Command{
		Name: "worker",
		Action: func(_ *cli.Context) error {
			// Init DB
			db, err := initMySQL(config)
			if err != nil {
				return err
			}
			repo := inframysql.NewProductRepository(db)

			svc := appservice.NewProductService(repo)

			activities := infratemporal.NewProductActivities(svc)

			// Init Temporal
			temporalHost := "temporal:7233"
			if config.TemporalHost != "" {
				temporalHost = config.TemporalHost
			}

			tClient, err := client.Dial(client.Options{
				HostPort: temporalHost,
			})
			if err != nil {
				return err
			}
			defer tClient.Close()

			w := worker.New(tClient, "product-task-queue", worker.Options{})

			// Explicitly register activities with string names used in Saga
			w.RegisterActivityWithOptions(activities.ReserveProduct, activity.RegisterOptions{Name: "ReserveProduct"})
			w.RegisterActivityWithOptions(activities.ReleaseProduct, activity.RegisterOptions{Name: "ReleaseProduct"})

			log.Println("Starting Product Temporal Worker...")
			return w.Run(worker.InterruptCh())
		},
	}
}
