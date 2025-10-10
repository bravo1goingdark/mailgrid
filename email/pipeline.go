package email

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SMTPPipeline provides optimized SMTP command pipelining
type SMTPPipeline struct {
	conn   net.Conn
	text   *textproto.Conn
	writer *bufio.Writer
	reader *bufio.Reader

	// Connection state
	ext     map[string]string
	mu      sync.Mutex
	pending int

	// Configuration
	maxPipeline   int
	flushInterval time.Duration
}

// NewSMTPPipeline creates a new SMTP pipeline with the given connection
func NewSMTPPipeline(conn net.Conn, maxPipeline int, flushInterval time.Duration) *SMTPPipeline {
	if maxPipeline <= 0 {
		maxPipeline = 100
	}
	if flushInterval <= 0 {
		flushInterval = 100 * time.Millisecond
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	return &SMTPPipeline{
		conn:          conn,
		text:          textproto.NewConn(conn),
		writer:        writer,
		reader:        reader,
		ext:           make(map[string]string),
		maxPipeline:   maxPipeline,
		flushInterval: flushInterval,
	}
}

// Pipeline represents a sequence of SMTP commands to be executed
type Pipeline struct {
	commands []string
	results  chan error
}

// NewPipeline creates a new command pipeline
func NewPipeline(size int) *Pipeline {
	return &Pipeline{
		commands: make([]string, 0, size),
		results:  make(chan error, size),
	}
}

// QueueCommand adds a command to the pipeline
func (p *Pipeline) QueueCommand(format string, args ...any) {
	cmd := fmt.Sprintf(format, args...)
	p.commands = append(p.commands, cmd)
}

// ExecutePipeline executes a batch of SMTP commands with pipelining
func (s *SMTPPipeline) ExecutePipeline(p *Pipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if pipelining is supported
	if _, ok := s.ext["PIPELINING"]; !ok {
		// Fall back to sequential execution
		return s.executeSequential(p)
	}

	// Split commands into batches
	var batches [][]string
	current := make([]string, 0, s.maxPipeline)

	for _, cmd := range p.commands {
		current = append(current, cmd)
		if len(current) >= s.maxPipeline {
			batches = append(batches, current)
			current = make([]string, 0, s.maxPipeline)
		}
	}
	if len(current) > 0 {
		batches = append(batches, current)
	}

	// Execute batches
	for _, batch := range batches {
		if err := s.executeBatch(batch); err != nil {
			return err
		}
	}

	return nil
}

// executeBatch executes a batch of commands using SMTP pipelining
func (s *SMTPPipeline) executeBatch(commands []string) error {
	// Write all commands
	for _, cmd := range commands {
		if _, err := fmt.Fprintf(s.writer, "%s\r\n", cmd); err != nil {
			return err
		}
		s.pending++
	}

	// Flush if needed
	if s.pending >= s.maxPipeline {
		if err := s.writer.Flush(); err != nil {
			return err
		}
	}

	// Read responses
	for i := 0; i < len(commands); i++ {
		code, msg, err := s.readResponse()
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("SMTP error %d: %s", code, msg)
		}
	}

	s.pending = 0
	return nil
}

// executeSequential executes commands sequentially when pipelining is not available
func (s *SMTPPipeline) executeSequential(p *Pipeline) error {
	for _, cmd := range p.commands {
		if _, err := fmt.Fprintf(s.writer, "%s\r\n", cmd); err != nil {
			return err
		}
		if err := s.writer.Flush(); err != nil {
			return err
		}

		code, msg, err := s.readResponse()
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("SMTP error %d: %s", code, msg)
		}
	}
	return nil
}

// readResponse reads an SMTP response
func (s *SMTPPipeline) readResponse() (code int, msg string, err error) {
	code = 0
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return 0, "", err
		}
		line = strings.TrimSpace(line)
		if len(line) < 4 || line[3] != ' ' && line[3] != '-' {
			continue
		}
		if code == 0 {
			code, err = strconv.Atoi(line[:3])
			if err != nil {
				return 0, "", err
			}
		}
		msg += line[4:]
		if line[3] == ' ' {
			break
		}
		msg += "\n"
	}
	return code, msg, nil
}

// StartTLS initiates a TLS session
func (s *SMTPPipeline) StartTLS(config *tls.Config) error {
	if _, ok := s.ext["STARTTLS"]; !ok {
		return fmt.Errorf("STARTTLS not supported")
	}

	if err := s.cmd(220, "STARTTLS"); err != nil {
		return err
	}

	tlsConn := tls.Client(s.conn, config)
	if err := tlsConn.Handshake(); err != nil {
		return err
	}

	s.conn = tlsConn
	s.text = textproto.NewConn(tlsConn)
	s.reader = bufio.NewReader(tlsConn)
	s.writer = bufio.NewWriter(tlsConn)

	return nil
}

// Auth performs SMTP authentication
func (s *SMTPPipeline) Auth(auth smtp.Auth) error {
	// This is a simplified auth implementation
	// For full implementation, we'd need to properly handle the auth protocol
	return fmt.Errorf("auth not implemented in pipeline yet")
}

// cmd sends a command and waits for the expected response code
func (s *SMTPPipeline) cmd(expectCode int, format string, args ...any) error {
	cmd := fmt.Sprintf(format, args...)
	if _, err := fmt.Fprintf(s.writer, "%s\r\n", cmd); err != nil {
		return err
	}
	if err := s.writer.Flush(); err != nil {
		return err
	}

	code, msg, err := s.readResponse()
	if err != nil {
		return err
	}
	if code != expectCode {
		return fmt.Errorf("unexpected response: %d %s", code, msg)
	}
	return nil
}

// Close closes the connection
func (s *SMTPPipeline) Close() error {
	if s.pending > 0 {
		_ = s.writer.Flush()
	}
	return s.conn.Close()
}
