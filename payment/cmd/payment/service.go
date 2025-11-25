package main

import (
	"context"
	"errors"
	"net"
	"net/http"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"payment/api/server/paymentinternal"
	appservice "payment/pkg/payment/app/service"
	"payment/pkg/payment/infrastructure/integrationevent"
	inframysql "payment/pkg/payment/infrastructure/mysql"
	"payment/pkg/payment/infrastructure/mysql/query"
	"payment/pkg/payment/infrastructure/transport"
	"payment/pkg/payment/infrastructure/transport/middlewares"
)

type serviceConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
}

func service(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "service",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[serviceConfig]()
			if err != nil {
				return err
			}

			closer := libio.NewMultiCloser()
			defer func() {
				err = errors.Join(err, closer.Close())
			}()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)
			databaseConnectionPool := mysql.NewConnectionPool(databaseConnector.TransactionalClient())

			libUoW := mysql.NewUnitOfWork(databaseConnectionPool, inframysql.NewRepositoryProvider)
			libLUow := mysql.NewLockableUnitOfWork(libUoW, mysql.NewLocker(databaseConnectionPool))
			uow := inframysql.NewUnitOfWork(libUoW)
			luow := inframysql.NewLockableUnitOfWork(libLUow)
			eventDispatcher := outbox.NewEventDispatcher(appID, integrationevent.TransportName, integrationevent.NewEventSerializer(), libUoW)

			userPublicAPIServer := transport.NewPaymentInternalAPI(
				query.NewPaymentQueryService(databaseConnector.TransactionalClient()),
				appservice.NewPaymentService(uow, luow, eventDispatcher),
				query.NewWalletQueryService(databaseConnector.TransactionalClient()),
				appservice.NewWalletService(uow, luow, eventDispatcher),
			)

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				listener, err := net.Listen("tcp", cnf.Service.GRPCAddress)
				if err != nil {
					return err
				}
				grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
					middlewares.NewGRPCLoggingMiddleware(logger),
				))
				paymentinternal.RegisterPaymentInternalAPIServer(grpcServer, userPublicAPIServer)
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, func(_ context.Context) error {
					grpcServer.GracefulStop()
					return nil
				})
				return grpcServer.Serve(listener)
			})
			errGroup.Go(func() error {
				router := mux.NewRouter()
				registerHealthcheck(router)
				// nolint:gosec
				server := http.Server{
					Addr:    cnf.Service.HTTPAddress,
					Handler: router,
				}
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, server.Shutdown)
				return server.ListenAndServe()
			})

			return errGroup.Wait()
		},
	}
}
