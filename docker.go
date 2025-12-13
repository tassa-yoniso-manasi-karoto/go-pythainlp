package pythainlp

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
	"github.com/tassa-yoniso-manasi-karoto/dockerutil"
)

const (
	defaultProjectName   = "pythainlp"
	defaultContainerName = "pythainlp-pythainlp-1"
	healthCheckPath      = "/health"
	serviceCheckInterval = 500 * time.Millisecond
	maxServiceWaitTime   = 480 * time.Second // account for first run = build take ~4min on low end CPU, low speed network

	// GHCR image for pre-built pythainlp container
	ghcrImage = "ghcr.io/tassa-yoniso-manasi-karoto/langkit-pythainlp:latest"
)

var (
	// Embed the service directory
	//go:embed service/*
	serviceFiles embed.FS

	// Embed requirements files
	//go:embed docker_light_requirements.txt
	lightRequirements []byte

	//go:embed docker_full_requirements.txt
	fullRequirements []byte

	// Default settings
	DefaultQueryTimeout   = 30 * time.Second
	DefaultDockerLogLevel = zerolog.TraceLevel
	
	// UseLightweightMode controls whether to use minimal dependencies (default: true)
	// Set to false before Init() if you need full PyThaiNLP capabilities
	UseLightweightMode = true
	
	// Logger for this package
	Logger = zerolog.Nop()

	// Package-level instance for backward compatibility
	instance       *PyThaiNLPManager
	instanceMu     sync.Mutex
	instanceClosed bool
)

// EnableDebugLogging enables debug logging for the package
func EnableDebugLogging() {
	Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly}).
		With().Timestamp().Logger()
}

// PyThaiNLPManager handles Docker lifecycle and service management for PyThaiNLP
type PyThaiNLPManager struct {
	docker                   *dockerutil.DockerManager
	logger                   *dockerutil.ContainerLogConsumer
	client                   *Client
	projectName              string
	containerName            string
	serviceURL               string
	servicePort              int
	QueryTimeout             time.Duration
	serviceReady             bool
	lightweightMode          bool
	downloadProgressCallback func(current, total int64, status string)
	mu                       sync.RWMutex
}

// ManagerOption defines function signature for options to configure PyThaiNLPManager
type ManagerOption func(*PyThaiNLPManager)

// WithQueryTimeout sets a custom query timeout
func WithQueryTimeout(timeout time.Duration) ManagerOption {
	return func(pm *PyThaiNLPManager) {
		pm.QueryTimeout = timeout
	}
}

// WithProjectName sets a custom project name for multiple instances
func WithProjectName(name string) ManagerOption {
	return func(pm *PyThaiNLPManager) {
		pm.projectName = name
		pm.containerName = name + "-pythainlp-1"
	}
}

// WithContainerName overrides the default container name
func WithContainerName(name string) ManagerOption {
	return func(pm *PyThaiNLPManager) {
		pm.containerName = name
	}
}

// WithLightweightMode sets whether to use lightweight mode (minimal dependencies)
func WithLightweightMode(lightweight bool) ManagerOption {
	return func(pm *PyThaiNLPManager) {
		pm.lightweightMode = lightweight
	}
}

// WithDownloadProgressCallback sets a callback for download progress during image pull
func WithDownloadProgressCallback(cb func(current, total int64, status string)) ManagerOption {
	return func(pm *PyThaiNLPManager) {
		pm.downloadProgressCallback = cb
	}
}

// ptr returns a pointer to the given string value
func ptr(s string) *string {
	return &s
}

// buildComposeProject creates the compose project definition for pythainlp
func buildComposeProject(dataDir string, port int) *types.Project {
	return &types.Project{
		Name: defaultProjectName,
		Services: types.Services{
			"pythainlp": {
				Name:       "pythainlp",
				Image:      ghcrImage,
				StdinOpen:  true,
				Tty:        true,
				WorkingDir: "/workspace",
				Environment: types.MappingWithEquals{
					"PYTHAINLP_DATA_DIR": ptr("/workspace/pythainlp-data"),
				},
				Volumes: []types.ServiceVolumeConfig{{
					Type:   types.VolumeTypeBind,
					Source: dataDir,
					Target: "/workspace",
				}},
				Ports: []types.ServicePortConfig{{
					Target:    uint32(port),
					Published: fmt.Sprintf("%d", port),
					Protocol:  "tcp",
				}},
			},
		},
	}
}

// NewManager creates a new PyThaiNLP manager instance
func NewManager(ctx context.Context, opts ...ManagerOption) (*PyThaiNLPManager, error) {
	// Enable Docker logging to stdout
	dockerutil.SetLogOutput(dockerutil.LogToStdout)

	manager := &PyThaiNLPManager{
		projectName:     defaultProjectName,
		containerName:   defaultContainerName,
		QueryTimeout:    DefaultQueryTimeout,
		lightweightMode: UseLightweightMode,
	}

	// Apply options
	for _, opt := range opts {
		opt(manager)
	}

	// Get XDG data directory for pythainlp
	dataDir := filepath.Join(xdg.ConfigHome, manager.projectName)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Allocate a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}
	manager.servicePort = listener.Addr().(*net.TCPAddr).Port
	listener.Close() // Release the port for later use

	Logger.Info().Int("port", manager.servicePort).Msg("Allocated port for PyThaiNLP service")

	// Build compose project
	project := buildComposeProject(dataDir, manager.servicePort)

	// Configure logging
	logConfig := dockerutil.LogConfig{
		Prefix:      manager.projectName,
		ShowService: true,
		ShowType:    true,
		LogLevel:    DefaultDockerLogLevel,
		InitMessage: "for more information",
	}

	logger := dockerutil.NewContainerLogConsumer(logConfig)

	// Configure Docker manager
	cfg := dockerutil.Config{
		ProjectName:      manager.projectName,
		Project:          project,
		RequiredServices: []string{"pythainlp"},
		LogConsumer:      logger,
		Timeout: dockerutil.Timeout{
			Create:   30 * time.Minute,
			Recreate: 60 * time.Minute,
			Start:    30 * time.Minute,
		},
		OnPullProgress: manager.downloadProgressCallback,
	}

	dockerManager, err := dockerutil.NewDockerManager(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker manager: %w", err)
	}

	manager.docker = dockerManager
	manager.logger = logger
	manager.serviceURL = fmt.Sprintf("http://localhost:%d", manager.servicePort)

	// Create HTTP client
	manager.client = NewClient(manager.serviceURL, manager.QueryTimeout)

	return manager, nil
}

// PullImage pre-pulls the GHCR image with progress tracking
func (pm *PyThaiNLPManager) PullImage(ctx context.Context) error {
	opts := dockerutil.DefaultPullOptions()
	if pm.downloadProgressCallback != nil {
		opts.OnProgress = pm.downloadProgressCallback
	}
	return dockerutil.PullImage(ctx, ghcrImage, opts)
}

// Init initializes the docker service and starts the Python server
func (pm *PyThaiNLPManager) Init(ctx context.Context) error {
	if err := pm.docker.Init(); err != nil {
		return fmt.Errorf("failed to initialize docker: %w", err)
	}

	// Start the Python service
	if err := pm.startService(ctx); err != nil {
		return fmt.Errorf("failed to start Python service: %w", err)
	}

	return nil
}

// InitRecreate removes existing containers then builds and starts new ones
func (pm *PyThaiNLPManager) InitRecreate(ctx context.Context, noCache bool) error {
	if noCache {
		if err := pm.docker.InitRecreateNoCache(); err != nil {
			return err
		}
	} else {
		if err := pm.docker.InitRecreate(); err != nil {
			return err
		}
	}

	// Start the Python service
	if err := pm.startService(ctx); err != nil {
		return fmt.Errorf("failed to start Python service: %w", err)
	}

	return nil
}

// copyRequirementsFile copies the appropriate requirements file based on lightweight mode
func (pm *PyThaiNLPManager) copyRequirementsFile() error {
	// Get the directory where dockerutil will look for files
	configDir, err := dockerutil.GetConfigDir(pm.projectName)
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Select the appropriate requirements file
	var requirements []byte
	if pm.lightweightMode {
		Logger.Info().Msg("Using lightweight requirements (no neural networks)")
		requirements = lightRequirements
	} else {
		Logger.Info().Msg("Using full requirements (includes neural networks)")
		requirements = fullRequirements
	}

	// Write as docker_requirements.txt
	targetPath := filepath.Join(configDir, "docker_requirements.txt")
	if err := os.WriteFile(targetPath, requirements, 0644); err != nil {
		return fmt.Errorf("failed to write requirements file: %w", err)
	}

	Logger.Debug().Str("path", targetPath).Bool("lightweight", pm.lightweightMode).Msg("Requirements file written")
	return nil
}

// startService copies the service files and starts the Python server
func (pm *PyThaiNLPManager) startService(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	Logger.Debug().Msg("Starting service...")

	// Get Docker client
	dockerClient, err := pm.docker.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	// Copy service files first
	Logger.Debug().Msg("Copying service files...")
	if err := pm.copyServiceFiles(ctx, dockerClient); err != nil {
		return fmt.Errorf("failed to copy service files: %w", err)
	}
	Logger.Debug().Msg("Service files copied successfully")

	// Check if service is already running
	Logger.Debug().Msg("Checking if service is already running...")
	if pm.isServiceRunning(ctx) {
		pm.serviceReady = true
		Logger.Debug().Msg("Service is already running")
		return nil
	}
	Logger.Debug().Msg("Service is not running, starting it...")

	// Start the service in a new bash session to avoid the interactive Python REPL
	startCmd := []string{
		"/bin/bash", "-c",
		"exec python -u /workspace/service/server.py",
	}

	execConfig := container.ExecOptions{
		Cmd:          startCmd,
		AttachStdout: false,
		AttachStderr: false,
		Detach:       true,
		Tty:          false,
		WorkingDir:   "/workspace",
	}

	exec, err := dockerClient.ContainerExecCreate(ctx, pm.containerName, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create service exec: %w", err)
	}

	Logger.Debug().Msg("Starting Python service exec...")
	if err := dockerClient.ContainerExecStart(ctx, exec.ID, container.ExecStartOptions{
		Detach: true,
		Tty:    false,
	}); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	Logger.Debug().Msg("Python service exec started")
	
	// Check if the file exists and see if Python started
	time.Sleep(2 * time.Second) // Give it a moment to start
	checkCmd := []string{"ps", "aux", "|", "grep", "server.py"}
	output, _ := pm.execCommand(ctx, dockerClient, checkCmd)
	Logger.Debug().Str("processes", string(output)).Msg("Process check")

	// Wait for service to be ready
	Logger.Debug().Msg("Waiting for service to be ready...")
	if err := pm.waitForService(ctx); err != nil {
		return fmt.Errorf("service failed to start: %w", err)
	}

	pm.serviceReady = true
	return nil
}

// copyServiceFiles copies the embedded service files into the container
func (pm *PyThaiNLPManager) copyServiceFiles(ctx context.Context, dockerClient *client.Client) error {
	// Read server.py from embedded files
	content, err := serviceFiles.ReadFile("service/server.py")
	if err != nil {
		return fmt.Errorf("failed to read server.py: %w", err)
	}

	// Replace port placeholder with actual port
	portStr := fmt.Sprintf("%d", pm.servicePort)
	modifiedContent := strings.ReplaceAll(string(content), "__PYTHAINLP_SERVICE_PORT__", portStr)
	
	// Verify replacement occurred
	if strings.Contains(modifiedContent, "__PYTHAINLP_SERVICE_PORT__") {
		return fmt.Errorf("failed to replace port placeholder in server.py")
	}

	// Create service directory in container
	mkdirCmd := []string{"mkdir", "-p", "/workspace/service"}
	if _, err := pm.execCommand(ctx, dockerClient, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create service directory: %w", err)
	}

	// Write server.py to container
	// Using a heredoc approach to write the file
	writeCmd := []string{
		fmt.Sprintf("cat > /workspace/service/server.py << 'EOF'\n%s\nEOF", modifiedContent),
	}
	if _, err := pm.execCommand(ctx, dockerClient, writeCmd); err != nil {
		return fmt.Errorf("failed to write server.py: %w", err)
	}

	// Make it executable
	chmodCmd := []string{"chmod", "+x", "/workspace/service/server.py"}
	if _, err := pm.execCommand(ctx, dockerClient, chmodCmd); err != nil {
		return fmt.Errorf("failed to chmod server.py: %w", err)
	}

	return nil
}

// execCommand executes a command in the container and returns the output
func (pm *PyThaiNLPManager) execCommand(ctx context.Context, dockerClient *client.Client, cmd []string) ([]byte, error) {
	// Use bash to execute commands since the container might have Python as the main process
	bashCmd := append([]string{"/bin/bash", "-c"}, strings.Join(cmd, " "))
	
	Logger.Trace().Strs("command", bashCmd).Msg("Executing command")
	
	execConfig := container.ExecOptions{
		Cmd:          bashCmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		WorkingDir:   "/workspace",
	}

	exec, err := dockerClient.ContainerExecCreate(ctx, pm.containerName, execConfig)
	if err != nil {
		return nil, err
	}

	resp, err := dockerClient.ContainerExecAttach(ctx, exec.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return nil, err
	}

	Logger.Trace().Str("output", string(output)).Msg("Command output")
	return output, nil
}

// isServiceRunning checks if the Python service is responding
func (pm *PyThaiNLPManager) isServiceRunning(ctx context.Context) bool {
	health, err := pm.client.Health(ctx)
	if err != nil {
		Logger.Trace().Err(err).Msg("Health check error")
		return false
	}
	Logger.Trace().Interface("response", health).Msg("Health check response")
	return err == nil && health.Status == "ready"
}

// waitForService waits for the Python service to be ready
func (pm *PyThaiNLPManager) waitForService(ctx context.Context) error {
	deadline := time.Now().Add(maxServiceWaitTime)
	
	attempt := 0
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(serviceCheckInterval):
			attempt++
			Logger.Trace().Int("attempt", attempt).Msg("Health check attempt")
			if pm.isServiceRunning(ctx) {
				Logger.Debug().Msg("Service is ready!")
				return nil
			}
			Logger.Trace().Msg("Service not ready yet")
		}
	}
	
	return fmt.Errorf("service failed to start within %v", maxServiceWaitTime)
}

// GetClient returns the HTTP client for making API calls
func (pm *PyThaiNLPManager) GetClient() *Client {
	return pm.client
}

// IsReady returns whether the service is ready to accept requests
func (pm *PyThaiNLPManager) IsReady() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.serviceReady
}

// IsLightweightMode returns whether the manager is using lightweight mode
func (pm *PyThaiNLPManager) IsLightweightMode() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.lightweightMode
}

// Stop stops the docker service
func (pm *PyThaiNLPManager) Stop(ctx context.Context) error {
	pm.mu.Lock()
	pm.serviceReady = false
	pm.mu.Unlock()
	
	return pm.docker.Stop()
}

// Close implements io.Closer
func (pm *PyThaiNLPManager) Close() error {
	pm.mu.Lock()
	pm.serviceReady = false
	pm.mu.Unlock()
	
	pm.logger.Close()
	return pm.docker.Close()
}

// Package-level functions for backward compatibility

// getOrCreateDefaultManager returns or creates the default manager instance
func getOrCreateDefaultManager(ctx context.Context) (*PyThaiNLPManager, error) {
	instanceMu.Lock()
	defer instanceMu.Unlock()

	if instance == nil || instanceClosed {
		mgr, err := NewManager(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create default manager: %w", err)
		}
		instance = mgr
		instanceClosed = false
	}

	return instance, nil
}

// Init initializes the default docker service
func Init() error {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return err
	}
	return mgr.Init(ctx)
}

// InitRecreate removes existing containers and creates new ones
func InitRecreate(noCache bool) error {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return err
	}
	return mgr.InitRecreate(ctx, noCache)
}

// Close closes the default instance
func Close() error {
	instanceMu.Lock()
	defer instanceMu.Unlock()

	if instance != nil {
		err := instance.Close()
		instanceClosed = true
		return err
	}
	return nil
}

// SetDefaultManager sets a custom manager as the package-level default instance.
// This allows providers that create their own manager to share it with code
// that uses package-level functions like SyllableTokenize().
//
// Use case: When a provider (e.g., translitkit's PyThaiNLPProvider) creates its
// own manager for lifecycle control, it can set that manager as the default so
// other code using package-level functions will reuse the same container.
func SetDefaultManager(mgr *PyThaiNLPManager) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	instance = mgr
	instanceClosed = false
}

// ClearDefaultManager clears the default manager reference.
// Call this when the manager is being closed to prevent stale references.
// This does NOT close the manager - the caller is responsible for that.
func ClearDefaultManager() {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	instance = nil
	instanceClosed = true
}