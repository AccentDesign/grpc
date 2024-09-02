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
	"time"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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

type TestSuite struct {
	suite.Suite
	emailServer *smtpmock.Server
	grpcConn    *grpc.ClientConn
}

func (suite *TestSuite) SetupSuite() {
	var err error
	var lis *bufconn.Listener
	suite.emailServer, lis = setupServer()
	suite.grpcConn, err = setupClientConn(lis)
	suite.NoError(err)
}

func setupServer() (*smtpmock.Server, *bufconn.Listener) {
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

func setupClientConn(lis *bufconn.Listener) (*grpc.ClientConn, error) {
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

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) waitForMessages() {
	// https://github.com/mocktools/go-smtp-mock/issues/181
	start := time.Now()
	for {
		if time.Since(start) > (5 * time.Second) {
			suite.Fail("Timeout waiting for messages")
		}
		if len(suite.emailServer.Messages()) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (suite *TestSuite) TestSendEmail_Validity() {
	client := pb.NewEmailServiceClient(suite.grpcConn)

	testErrorCases := []struct {
		desc          string
		info          *pb.EmailInfo
		attachment    *pb.Attachment
		expectedError error
	}{
		{"no email info", nil, &pb.Attachment{Filename: "test.txt", Data: []byte("123"), ContentType: "text/plain"}, status.Error(codes.InvalidArgument, "EmailInfo not found in stream")},
		{"missing from address", &pb.EmailInfo{}, nil, status.Error(codes.InvalidArgument, "from_address is required")},
		{"missing to address", &pb.EmailInfo{FromAddress: "from@mail.com"}, nil, status.Error(codes.InvalidArgument, "to_address is required")},
		{"missing subject", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com"}, nil, status.Error(codes.InvalidArgument, "subject is required")},
		{"missing html or plain", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi"}, nil, status.Error(codes.InvalidArgument, "plain_text or html is required")},
		{"missing filename", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi", PlainText: "hi"}, &pb.Attachment{}, status.Error(codes.InvalidArgument, "filename is required")},
		{"missing data", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi", PlainText: "hi"}, &pb.Attachment{Filename: "test.txt"}, status.Error(codes.InvalidArgument, "data is required")},
		{"missing content type", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi", PlainText: "hi"}, &pb.Attachment{Filename: "test.txt", Data: []byte("123")}, status.Error(codes.InvalidArgument, "content_type is required")},
	}

	for _, tc := range testErrorCases {
		suite.Run(tc.desc, func() {
			stream, err := client.SendEmail(context.Background())
			suite.NoError(err)

			if tc.info != nil {
				err = stream.Send(&pb.EmailRequest{Payload: &pb.EmailRequest_EmailInfo{EmailInfo: tc.info}})
				suite.NoError(err)
			}

			if tc.attachment != nil {
				err = stream.Send(&pb.EmailRequest{Payload: &pb.EmailRequest_Attachment{Attachment: tc.attachment}})
				suite.NoError(err)
			}

			_, err = stream.CloseAndRecv()
			suite.EqualError(err, tc.expectedError.Error())
		})
	}

	testOkCases := []struct {
		desc       string
		info       *pb.EmailInfo
		attachment *pb.Attachment
	}{
		{"plain", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi", PlainText: "Hi"}, nil},
		{"html", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi", Html: "<p>Hi</p>"}, nil},
		{"both", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi", PlainText: "Hi", Html: "<p>Hi</p>"}, nil},
		{"with attachment", &pb.EmailInfo{FromAddress: "from@mail.com", ToAddress: "to@mail.com", Subject: "Hi", PlainText: "Hi", Html: "<p>Hi</p>"}, &pb.Attachment{Filename: "test.txt", Data: []byte("123"), ContentType: "text/plain"}},
	}

	for _, tc := range testOkCases {
		suite.Run(tc.desc, func() {
			stream, err := client.SendEmail(context.Background())
			suite.NoError(err)

			if tc.info != nil {
				err = stream.Send(&pb.EmailRequest{Payload: &pb.EmailRequest_EmailInfo{EmailInfo: tc.info}})
				suite.NoError(err)
			}

			if tc.attachment != nil {
				err = stream.Send(&pb.EmailRequest{Payload: &pb.EmailRequest_Attachment{Attachment: tc.attachment}})
				suite.NoError(err)
			}

			_, err = stream.CloseAndRecv()
			suite.NoError(err)
		})
	}
}

func (suite *TestSuite) TestSendEmail_WithAttachments() {
	client := pb.NewEmailServiceClient(suite.grpcConn)

	// Send a test email.
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	_, err := encoder.Write([]byte("This is a test attachment"))
	suite.NoError(err)
	err = encoder.Close()
	suite.NoError(err)

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
	suite.NoError(err)

	err = stream.Send(emailRequest1)
	suite.NoError(err)

	err = stream.Send(emailRequest2)
	suite.NoError(err)

	response, err := stream.CloseAndRecv()
	suite.NoError(err)

	// Check that the email was sent successfully.
	suite.True(response.Success)
	suite.Equal("Email sent successfully", response.Message)

	suite.waitForMessages()

	// Verify that the email was sent correctly.
	messages := suite.emailServer.Messages()
	message := last(messages)

	suite.Equal("MAIL FROM:<from@example.com>", message.MailfromRequest())

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

	suite.Equal(expectedMsg.String(), message.MsgRequest())

}

func (suite *TestSuite) TestSendEmail_WithoutAttachments() {
	client := pb.NewEmailServiceClient(suite.grpcConn)

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
	suite.NoError(err)

	err = stream.Send(emailRequest1)
	suite.NoError(err)

	response, err := stream.CloseAndRecv()
	suite.NoError(err)

	// Check that the email was sent successfully.
	suite.True(response.Success)
	suite.Equal("Email sent successfully", response.Message)

	suite.waitForMessages()

	// Verify that the email was sent correctly.
	messages := suite.emailServer.Messages()
	message := last(messages)

	suite.Equal("MAIL FROM:<from@example.com>", message.MailfromRequest())

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

	suite.Equal(expectedMsg.String(), message.MsgRequest())

}

func (suite *TestSuite) TestSendEmail_PlainOnly() {
	client := pb.NewEmailServiceClient(suite.grpcConn)

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
	suite.NoError(err)

	err = stream.Send(emailRequest1)
	suite.NoError(err)

	response, err := stream.CloseAndRecv()
	suite.NoError(err)

	// Check that the email was sent successfully.
	suite.True(response.Success)
	suite.Equal("Email sent successfully", response.Message)

	suite.waitForMessages()

	// Verify that the email was sent correctly.
	messages := suite.emailServer.Messages()
	message := last(messages)

	suite.Equal("MAIL FROM:<from@example.com>", message.MailfromRequest())

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

	suite.Equal(expectedMsg.String(), message.MsgRequest())

}

func (suite *TestSuite) TestSendEmail_HtmlOnly() {
	client := pb.NewEmailServiceClient(suite.grpcConn)

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
	suite.NoError(err)

	err = stream.Send(emailRequest1)
	suite.NoError(err)

	response, err := stream.CloseAndRecv()
	suite.NoError(err)

	// Check that the email was sent successfully.
	suite.True(response.Success)
	suite.Equal("Email sent successfully", response.Message)

	suite.waitForMessages()

	// Verify that the email was sent correctly.
	messages := suite.emailServer.Messages()
	message := last(messages)

	suite.Equal("MAIL FROM:<from@example.com>", message.MailfromRequest())

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

	suite.Equal(expectedMsg.String(), message.MsgRequest())

}

func (suite *TestSuite) TestSendEmail_MissingEmailInfo() {
	client := pb.NewEmailServiceClient(suite.grpcConn)

	stream, err := client.SendEmail(context.Background())
	suite.NoError(err)

	_, err = stream.CloseAndRecv()
	expected := status.Error(codes.InvalidArgument, "EmailInfo not found in stream")
	suite.EqualError(err, expected.Error())
}
