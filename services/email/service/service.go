package service

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"strings"

	"github.com/asaskevich/govalidator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/accentdesign/grpc/services/email/internal"
	pb "github.com/accentdesign/grpc/services/email/pkg/api/email"
)

type EmailServer struct {
	pb.UnimplementedEmailServiceServer
	boundaryGenerator internal.BoundaryGenerator
	host              string
	port              int64
	username          string
	password          string
	startTLS          bool

	auth smtp.Auth
	conn *smtp.Client
}

func NewEmailServer(host string, port int64, username string, password string, startTLS bool) (*EmailServer, error) {
	s := &EmailServer{
		boundaryGenerator: &internal.DefaultBoundaryGenerator{},
		host:              host,
		port:              port,
		username:          username,
		password:          password,
		startTLS:          startTLS,
	}
	if err := s.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize email server: %v", err)
	}
	return s, nil
}

func (s *EmailServer) init() error {
	var err error
	s.auth, s.conn, err = s.setupSMTPConnection()
	if err != nil {
		return fmt.Errorf("error setting up SMTP connection: %v", err)
	}
	return nil
}

func (s *EmailServer) SetBoundaryGenerator(boundaryGenerator internal.BoundaryGenerator) {
	s.boundaryGenerator = boundaryGenerator
}

func (s *EmailServer) SendEmail(stream pb.EmailService_SendEmailServer) error {
	var emailInfo *pb.EmailInfo
	var attachments []*pb.Attachment

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			if emailInfo == nil {
				return status.Error(codes.InvalidArgument, "EmailInfo not found in stream")
			}
			sendErr := s.send(emailInfo, attachments)
			success := sendErr == nil
			message := ""
			if success {
				message = "Email sent successfully"
			} else {
				message = sendErr.Error()
			}
			return stream.SendAndClose(&pb.EmailResponse{
				Success: success,
				Message: message,
			})
		}
		if err != nil {
			return err
		}

		switch payload := req.Payload.(type) {
		case *pb.EmailRequest_EmailInfo:
			if govalidator.IsNull(payload.EmailInfo.GetFromAddress()) {
				return status.Error(codes.InvalidArgument, "from_address is required")
			}
			if govalidator.IsNull(payload.EmailInfo.GetToAddress()) {
				return status.Error(codes.InvalidArgument, "to_address is required")
			}
			if govalidator.IsNull(payload.EmailInfo.GetSubject()) {
				return status.Error(codes.InvalidArgument, "subject is required")
			}
			if govalidator.IsNull(payload.EmailInfo.GetPlainText()) && govalidator.IsNull(payload.EmailInfo.GetHtml()) {
				return status.Error(codes.InvalidArgument, "plain_text or html is required")
			}
			emailInfo = payload.EmailInfo
		case *pb.EmailRequest_Attachment:
			if govalidator.IsNull(payload.Attachment.GetFilename()) {
				return status.Error(codes.InvalidArgument, "filename is required")
			}
			if len(payload.Attachment.GetData()) == 0 {
				return status.Error(codes.InvalidArgument, "data is required")
			}
			if govalidator.IsNull(payload.Attachment.GetContentType()) {
				return status.Error(codes.InvalidArgument, "content_type is required")
			}
			attachments = append(attachments, payload.Attachment)
		default:
			return status.Errorf(codes.InvalidArgument, "unknown payload received: %T", payload)
		}
	}
}

func (s *EmailServer) send(info *pb.EmailInfo, attachments []*pb.Attachment) error {
	log.Printf("EmailInfo: %v", info)
	log.Printf("Attachments: %v", len(attachments))

	serverAddr := fmt.Sprintf("%s:%d", s.host, s.port)
	from := info.GetFromAddress()
	to := []string{info.GetToAddress()}
	subject := info.GetSubject()
	plainText := info.GetPlainText()
	htmlBody := info.GetHtml()

	var err error
	s.auth, s.conn, err = s.setupSMTPConnection()
	if err != nil {
		return err
	}

	boundary, err := s.boundaryGenerator.GetBoundary()
	if err != nil {
		return err
	}

	message, err := createEmailMessage(from, to, subject, plainText, htmlBody, boundary, attachments)
	if err != nil {
		return err
	}

	return smtp.SendMail(serverAddr, s.auth, from, to, message.Bytes())
}

func (s *EmailServer) setupSMTPConnection() (smtp.Auth, *smtp.Client, error) {
	serverAddr := fmt.Sprintf("%s:%d", s.host, s.port)

	conn, err := smtp.Dial(serverAddr)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if err != nil {
			if closeErr := conn.Close(); closeErr != nil {
				log.Printf("Error closing SMTP connection after setup error: %v", closeErr)
			}
		}
	}()

	if s.startTLS {
		tlsConfig := &tls.Config{ServerName: s.host}
		if err := conn.StartTLS(tlsConfig); err != nil {
			return nil, nil, err
		}
	}

	var auth smtp.Auth
	if s.username != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
		if err := conn.Auth(auth); err != nil {
			return nil, nil, err
		}
	}

	return auth, conn, nil
}

func createEmailMessage(from string, to []string, subject, plainText, htmlBody, boundary string, attachments []*pb.Attachment) (*bytes.Buffer, error) {
	message := bytes.NewBuffer(nil)
	message.WriteString(fmt.Sprintf("From: %s\r\n", from))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=outer-%s\r\n\r\n", boundary))

	message.WriteString(fmt.Sprintf("--outer-%s\r\n", boundary))
	message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=inner-%s\r\n\r\n", boundary))

	if plainText != "" {
		addPlainTextBody(message, plainText, boundary)
	}

	if htmlBody != "" {
		addHTMLBody(message, htmlBody, boundary)
	}

	message.WriteString(fmt.Sprintf("--inner-%s--\r\n\r\n", boundary))

	for _, attachment := range attachments {
		if err := addAttachment(message, attachment, boundary); err != nil {
			return nil, err
		}
	}

	message.WriteString(fmt.Sprintf("--outer-%s--", boundary))

	return message, nil
}

func addPlainTextBody(message *bytes.Buffer, plainText, boundary string) {
	message.WriteString(fmt.Sprintf("--inner-%s\r\n", boundary))
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	message.WriteString(plainText + "\r\n\r\n")
}

func addHTMLBody(message *bytes.Buffer, htmlBody, boundary string) {
	message.WriteString(fmt.Sprintf("--inner-%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")
}

func addAttachment(message *bytes.Buffer, attachment *pb.Attachment, boundary string) error {
	filename := attachment.GetFilename()
	contentType := attachment.GetContentType()
	data := attachment.GetData()

	message.WriteString(fmt.Sprintf("--outer-%s\r\n", boundary))
	message.WriteString(fmt.Sprintf("Content-Type: %s; name=%s\r\n", contentType, filename))
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n\r\n", filename))

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	message.Write(encoded)
	message.WriteString("\r\n\r\n")

	return nil
}
