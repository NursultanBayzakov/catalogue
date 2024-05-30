package grpcapp

import (
	catalogueGrpc "catalogue-service/internal/grpc/catalogue"
	"context"
	"fmt"
	ssov1 "github.com/bxiit/protos/gen/go/sso"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

var authServiceClient ssov1.AuthClient

func New(
	log *slog.Logger,
	catalogueService catalogueGrpc.Catalogue,
	port int,
) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			//logging.StartCall, logging.FinishCall,
			logging.PayloadReceived, logging.PayloadSent,
		),
		// Add any other option (check functions starting with logging.With).
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
	), grpc.UnaryInterceptor(AuthInterceptor))

	catalogueGrpc.Register(gRPCServer, catalogueService)

	// conn to sso
	connectToSsoService()
	return &App{
		log:        log,
		port:       port,
		grpcServer: gRPCServer,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run runs gRPC server.
func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/catalogue.CatalogueService/CreateItem" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	claims, err := DecodeToken("catalogue-secret", tkn[0])
	if err != nil {
		log.Printf("failed to decode jwt %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "authentication failed")
	}

	isAdminRequest := &ssov1.IsAdminRequest{UserId: claims.UID}
	isAdminResponse, err := authServiceClient.IsAdmin(ctx, isAdminRequest)
	if err != nil {
		log.Printf("permissions fail %v", err)
		return nil, status.Errorf(codes.Internal, "permission failed")
	}

	if !isAdminResponse.IsAdmin {
		return nil, status.Errorf(codes.Internal, "permission failed")
	}

	return handler(ctx, req)
}

func connectToSsoService() {
	conn, err := grpc.NewClient("0.0.0.0:44044", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	authServiceClient = ssov1.NewAuthClient(conn)
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.grpcServer.GracefulStop()
}

type TokenClaims struct {
	UID   int64  `json:"uid"`
	Email string `json:"email"`
	AppID int    `json:"app_id"`
	jwt.MapClaims
}

func DecodeToken(appSecret string, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
