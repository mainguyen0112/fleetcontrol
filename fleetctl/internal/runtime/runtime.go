package runtime

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

const (
	DefaultServer      = "http://localhost:8080"
	DefaultHTTPTimeout = 15 * time.Second
)

type Config struct {
	Server string
	Token  string
}

type Options struct {
	ConfigPath string
	Server     string
	HTTPClient *http.Client
}

func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".fleetctl", "config.yaml"), nil
}

func (o *Options) path() (string, error) {
	if o.ConfigPath != "" {
		return o.ConfigPath, nil
	}
	return DefaultConfigPath()
}

func (o *Options) Load() (Config, error) {
	cfg := Config{Server: DefaultServer}
	path, err := o.path()
	if err != nil {
		return cfg, err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		if o.Server != "" {
			cfg.Server = o.Server
		}
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.TrimSpace(parts[1])
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		switch strings.TrimSpace(parts[0]) {
		case "server":
			cfg.Server = value
		case "token":
			cfg.Token = value
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, err
	}
	if o.Server != "" {
		cfg.Server = o.Server
	}
	return cfg, nil
}

func (o *Options) Save(cfg Config) error {
	path, err := o.path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content := fmt.Sprintf("server: %s\ntoken: %s\n", strconv.Quote(cfg.Server), strconv.Quote(cfg.Token))
	return os.WriteFile(path, []byte(content), 0o600)
}

func (o *Options) Client(requireAuth bool) (*gen.ClientWithResponses, []gen.RequestEditorFn, error) {
	cfg, err := o.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	if requireAuth && cfg.Token == "" {
		return nil, nil, fmt.Errorf("not logged in; run fleetctl login")
	}
	httpClient := o.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	client, err := gen.NewClientWithResponses(strings.TrimRight(cfg.Server, "/"), gen.WithHTTPClient(httpClient))
	if err != nil {
		return nil, nil, fmt.Errorf("create API client: %w", err)
	}
	if cfg.Token == "" {
		return client, nil, nil
	}
	editor := func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
		return nil
	}
	return client, []gen.RequestEditorFn{editor}, nil
}
