package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/accentdesign/grpc/services/email/pkg/api/email"
	"github.com/accentdesign/grpc/services/email/service"
)

var (
	helpFlag         = flag.Bool("help", false, "Display help information")
	enableReflection = flag.Bool("reflection", false, "Enable reflection")
	port             = flag.Int("port", 50051, "The server port")
	smtpHost         = os.Getenv("SMTP_HOST")
	smtpPort         = os.Getenv("SMTP_PORT")
	smtpUsername     = os.Getenv("SMTP_USERNAME")
	smtpPassword     = os.Getenv("SMTP_PASSWORD")
	smtpStartTLS     = os.Getenv("SMTP_STARTTLS")
)

func displayHelp() {
	flag.PrintDefaults()
	fmt.Println("Environment variables:")
}

func main() {
	// get runtime flags
	flag.Parse()

	// display help
	if *helpFlag {
		displayHelp()
		os.Exit(0)
	}

	// log errors
	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			log.Printf("method: %q error: %s", info.FullMethod, err)
		}
		return resp, err
	}

	// create the server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(errHandler),
	)

	// ensure env vars convert to their proper types
	sTLS := false
	if smtpStartTLS != "" {
		var err error
		sTLS, err = strconv.ParseBool(smtpStartTLS)
		if err != nil {
			log.Fatalf("Invalid value for SMTP_STARTTLS: %v", err.Error())
		}
	}
	sPort, err := strconv.ParseInt(smtpPort, 10, 64)
	if err != nil {
		log.Fatalf("Invalid value for SMTP_PORT: %v", err.Error())
	}

	// define the service
	log.Print("checking email server settings..")
	emailService, err := service.NewEmailServer(smtpHost, sPort, smtpUsername, smtpPassword, sTLS)
	if err != nil {
		log.Fatalf("failed to initialize email service: %v", err)
	}

	// register the services
	pb.RegisterEmailServiceServer(grpcServer, emailService)

	// enable reflection
	if *enableReflection {
		reflection.Register(grpcServer)
		log.Print("reflection enabled")
	}

	// configure the listen address
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// serve
	log.Printf("server listening at %v", listen.Addr())
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
