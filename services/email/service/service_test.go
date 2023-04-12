package service_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"testing"
	"text/template"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/accentdesign/grpc/services/email/pkg/api/email"
	"github.com/accentdesign/grpc/services/email/service"
)

func last[E any](s []E) E {
	return s[len(s)-1]
}

type MockBoundaryGenerator struct{}

func (g *MockBoundaryGenerator) GetBoundary() (string, error) {
	return "mocked-id", nil
}

func setupMockServer() (*smtpmock.Server, *bufconn.Listener) {
	mockServer := smtpmock.New(smtpmock.ConfigurationAttr{})
	if err := mockServer.Start(); err != nil {
		fmt.Println(err)
	}

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()

	hostAddress, portNumber := "127.0.0.1", mockServer.PortNumber()
	emailServer, sErr := service.NewEmailServer(hostAddress, int64(portNumber), "", "", false)
	if sErr != nil {
		log.Fatalf("Error defining service: %v", sErr)
	}
	emailServer.SetBoundaryGenerator(&MockBoundaryGenerator{})
	pb.RegisterEmailServiceServer(srv, emailServer)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("Error starting gRPC server: %v", err)
		}
	}()

	return mockServer, lis
}

func setupGRPCClient(lis *bufconn.Listener) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(context.Background(), "",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
	if err != nil {
		return nil, err
	}
	runtime.SetFinalizer(conn, func(interface{}) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Could not close connection: %v", err.Error())
		}
	})
	return conn, nil
}

func TestSendEmail_WithAttachments(t *testing.T) {
	mockServer, lis := setupMockServer()
	defer mockServer.Stop()

	conn, err := setupGRPCClient(lis)
	if err != nil {
		t.Fatalf("Error setting up gRPC client: %v", err)
	}

	client := pb.NewEmailServiceClient(conn)

	// Send a test email.
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte("This is a test attachment"))
	encoder.Close()

	emailRequest1 := &pb.EmailRequest{
		Payload: &pb.EmailRequest_EmailInfo{
			EmailInfo: &pb.EmailInfo{
				FromAddress: "from@example.com",
				ToAddress:   "to@example.com",
				Subject:     "Test email",
				PlainText:   "This is a test email",
				Html:        "<html><body><p>This is a test email.</p></body></html>",
			},
		},
	}
	emailRequest2 := &pb.EmailRequest{
		Payload: &pb.EmailRequest_Attachment{
			Attachment: &pb.Attachment{
				Filename:    "test.txt",
				Data:        buf.Bytes(),
				ContentType: "text/plain",
			},
		},
	}

	stream, err := client.SendEmail(context.Background())
	if err != nil {
		t.Fatalf("Error sending email: %v", err)
	}

	if err := stream.Send(emailRequest1); err != nil {
		t.Fatalf("Error sending email request: %v", err)
	}

	if err := stream.Send(emailRequest2); err != nil {
		t.Fatalf("Error sending email request: %v", err)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("Error receiving response: %v", err)
	}

	// Check that the email was sent successfully.
	assert.True(t, response.Success)
	assert.Equal(t, "Email sent successfully", response.Message)

	// Verify that the email was sent correctly.
	messages := mockServer.Messages()
	message := last(messages)

	assert.Equal(t, "MAIL FROM:<from@example.com>", message.MailfromRequest())

	data := map[string]interface{}{
		"from":         "from@example.com",
		"to":           "to@example.com",
		"subject":      "Test email",
		"plain":        "This is a test email",
		"html":         "<html><body><p>This is a test email.</p></body></html>",
		"file_name":    "test.txt",
		"content_type": "text/plain",
		"data":         base64.StdEncoding.EncodeToString(buf.Bytes()),
	}
	tmpl, err := template.New("").Parse(strings.ReplaceAll(`From: {{.from}}
To: {{.to}}
Subject: {{.subject}}
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=outer-mocked-id

--outer-mocked-id
Content-Type: multipart/alternative; boundary=inner-mocked-id

--inner-mocked-id
Content-Type: text/plain; charset=utf-8

{{.plain}}

--inner-mocked-id
Content-Type: text/html; charset=utf-8

{{.html}}

--inner-mocked-id--

--outer-mocked-id
Content-Type: {{.content_type}}; name={{.file_name}}
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename={{.file_name}}

{{.data}}

--outer-mocked-id--
`, "\n", "\r\n"))
	if err != nil {
		panic(err)
	}
	var expectedMsg bytes.Buffer
	if err := tmpl.Execute(&expectedMsg, data); err != nil {
		panic(err)
	}

	assert.Equal(t, expectedMsg.String(), message.MsgRequest())

}

func TestSendEmail_WithoutAttachments(t *testing.T) {
	mockServer, lis := setupMockServer()
	defer mockServer.Stop()

	conn, err := setupGRPCClient(lis)
	if err != nil {
		t.Fatalf("Error setting up gRPC client: %v", err)
	}

	client := pb.NewEmailServiceClient(conn)

	// Send a test email.
	emailRequest1 := &pb.EmailRequest{
		Payload: &pb.EmailRequest_EmailInfo{
			EmailInfo: &pb.EmailInfo{
				FromAddress: "from@example.com",
				ToAddress:   "to@example.com",
				Subject:     "Test email",
				PlainText:   "This is a test email",
				Html:        "<html><body><p>This is a test email.</p></body></html>",
			},
		},
	}

	stream, err := client.SendEmail(context.Background())
	if err != nil {
		t.Fatalf("Error sending email: %v", err)
	}

	if err := stream.Send(emailRequest1); err != nil {
		t.Fatalf("Error sending email request: %v", err)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("Error receiving response: %v", err)
	}

	// Check that the email was sent successfully.
	assert.True(t, response.Success)
	assert.Equal(t, "Email sent successfully", response.Message)

	// Verify that the email was sent correctly.
	messages := mockServer.Messages()
	message := last(messages)

	assert.Equal(t, "MAIL FROM:<from@example.com>", message.MailfromRequest())

	data := map[string]interface{}{
		"from":    "from@example.com",
		"to":      "to@example.com",
		"subject": "Test email",
		"plain":   "This is a test email",
		"html":    "<html><body><p>This is a test email.</p></body></html>",
	}
	tmpl, err := template.New("").Parse(strings.ReplaceAll(`From: {{.from}}
To: {{.to}}
Subject: {{.subject}}
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=outer-mocked-id

--outer-mocked-id
Content-Type: multipart/alternative; boundary=inner-mocked-id

--inner-mocked-id
Content-Type: text/plain; charset=utf-8

{{.plain}}

--inner-mocked-id
Content-Type: text/html; charset=utf-8

{{.html}}

--inner-mocked-id--

--outer-mocked-id--
`, "\n", "\r\n"))
	if err != nil {
		panic(err)
	}
	var expectedMsg bytes.Buffer
	if err := tmpl.Execute(&expectedMsg, data); err != nil {
		panic(err)
	}

	assert.Equal(t, expectedMsg.String(), message.MsgRequest())

}

func TestSendEmail_PlainOnly(t *testing.T) {
	mockServer, lis := setupMockServer()
	defer mockServer.Stop()

	conn, err := setupGRPCClient(lis)
	if err != nil {
		t.Fatalf("Error setting up gRPC client: %v", err)
	}

	client := pb.NewEmailServiceClient(conn)

	// Send a test email.
	emailRequest1 := &pb.EmailRequest{
		Payload: &pb.EmailRequest_EmailInfo{
			EmailInfo: &pb.EmailInfo{
				FromAddress: "from@example.com",
				ToAddress:   "to@example.com",
				Subject:     "Test email",
				PlainText:   "This is a test email",
			},
		},
	}

	stream, err := client.SendEmail(context.Background())
	if err != nil {
		t.Fatalf("Error sending email: %v", err)
	}

	if err := stream.Send(emailRequest1); err != nil {
		t.Fatalf("Error sending email request: %v", err)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("Error receiving response: %v", err)
	}

	// Check that the email was sent successfully.
	assert.True(t, response.Success)
	assert.Equal(t, "Email sent successfully", response.Message)

	// Verify that the email was sent correctly.
	messages := mockServer.Messages()
	message := last(messages)

	assert.Equal(t, "MAIL FROM:<from@example.com>", message.MailfromRequest())

	data := map[string]interface{}{
		"from":    "from@example.com",
		"to":      "to@example.com",
		"subject": "Test email",
		"plain":   "This is a test email",
	}
	tmpl, err := template.New("").Parse(strings.ReplaceAll(`From: {{.from}}
To: {{.to}}
Subject: {{.subject}}
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=outer-mocked-id

--outer-mocked-id
Content-Type: multipart/alternative; boundary=inner-mocked-id

--inner-mocked-id
Content-Type: text/plain; charset=utf-8

{{.plain}}

--inner-mocked-id--

--outer-mocked-id--
`, "\n", "\r\n"))
	if err != nil {
		panic(err)
	}
	var expectedMsg bytes.Buffer
	if err := tmpl.Execute(&expectedMsg, data); err != nil {
		panic(err)
	}

	assert.Equal(t, expectedMsg.String(), message.MsgRequest())

}

func TestSendEmail_HtmlOnly(t *testing.T) {
	mockServer, lis := setupMockServer()
	defer mockServer.Stop()

	conn, err := setupGRPCClient(lis)
	if err != nil {
		t.Fatalf("Error setting up gRPC client: %v", err)
	}

	client := pb.NewEmailServiceClient(conn)

	// Send a test email.
	emailRequest1 := &pb.EmailRequest{
		Payload: &pb.EmailRequest_EmailInfo{
			EmailInfo: &pb.EmailInfo{
				FromAddress: "from@example.com",
				ToAddress:   "to@example.com",
				Subject:     "Test email",
				Html:        "<html><body><p>This is a test email.</p></body></html>",
			},
		},
	}

	stream, err := client.SendEmail(context.Background())
	if err != nil {
		t.Fatalf("Error sending email: %v", err)
	}

	if err := stream.Send(emailRequest1); err != nil {
		t.Fatalf("Error sending email request: %v", err)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("Error receiving response: %v", err)
	}

	// Check that the email was sent successfully.
	assert.True(t, response.Success)
	assert.Equal(t, "Email sent successfully", response.Message)

	// Verify that the email was sent correctly.
	messages := mockServer.Messages()
	message := last(messages)

	assert.Equal(t, "MAIL FROM:<from@example.com>", message.MailfromRequest())

	data := map[string]interface{}{
		"from":    "from@example.com",
		"to":      "to@example.com",
		"subject": "Test email",
		"html":    "<html><body><p>This is a test email.</p></body></html>",
	}
	tmpl, err := template.New("").Parse(strings.ReplaceAll(`From: {{.from}}
To: {{.to}}
Subject: {{.subject}}
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=outer-mocked-id

--outer-mocked-id
Content-Type: multipart/alternative; boundary=inner-mocked-id

--inner-mocked-id
Content-Type: text/html; charset=utf-8

{{.html}}

--inner-mocked-id--

--outer-mocked-id--
`, "\n", "\r\n"))
	if err != nil {
		panic(err)
	}
	var expectedMsg bytes.Buffer
	if err := tmpl.Execute(&expectedMsg, data); err != nil {
		panic(err)
	}

	assert.Equal(t, expectedMsg.String(), message.MsgRequest())

}
