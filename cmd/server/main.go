package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"

	pb "product-catalog-service/gen/go/product/v1"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
	"product-catalog-service/internal/transport/grpc/product"
)

func main() {
	// 1. Setup Signal Handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 2. Infrastructure Setup (Spanner)
	database := os.Getenv("SPANNER_DATABASE")
	if database == "" {
		database = "projects/test-project/instances/test-instance/databases/test-database"
	}
	
	client, err := spanner.NewClient(ctx, database)
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer client.Close()

	// 3. Dependency Injection
	pr := repo.NewProductRepository(client)
	or := repo.NewOutboxRepository()
	cm := committer.NewCommitter(client)
	cl := clock.RealClock{}

	// 4. Transport Setup (Injecting dependencies into the gRPC handler)
	productHandler := product.NewHandler(
		client,
		pr,
		or,
		cm,
		cl,
	)

	srv := grpc.NewServer()
	pb.RegisterProductServiceServer(srv, productHandler)

	// 5. Run Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("Starting gRPC server on :50051...")
	go func() {
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 6. Wait for Shutdown
	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	srv.GracefulStop()
	log.Println("Server stopped.")
}