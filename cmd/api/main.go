package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/banks/simulator"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/config"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/payments"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

//	@title			Payment Gateway Challenge Go
//	@description	Interview challenge for building a Payment Gateway - Go version

//	@host		localhost:8090
//	@BasePath	/

// @securityDefinitions.basic	BasicAuth
func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading environment variables: %v", err)
	}

	fmt.Printf("version %s, commit %s, built at %s\n", version, commit, date)
	docs.SwaggerInfo.Version = version

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Printf("sigterm/interrupt signal\n")
		cancel()
	}()

	defer func() {
		// recover after panic
		if x := recover(); x != nil {
			fmt.Printf("run time panic:\n%v\n", x)
			panic(x)
		}
	}()

	paymentsRepository := repository.NewPaymentsRepositoryInMemory()

	bankSimulator := simulator.NewClient(conf.BankSimulator.URL, nil)
	paymentsSvc := payments.NewService(paymentsRepository, bankSimulator)

	paymentsHandler := api.NewPaymentsHandler(paymentsSvc)
	api := api.New(paymentsHandler)

	if err := api.Run(ctx, ":"+conf.App.APIPort); err != nil {
		log.Fatalf("error setup the API: %v", err)
	}
}
