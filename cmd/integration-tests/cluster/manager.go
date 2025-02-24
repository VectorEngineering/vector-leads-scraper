package cluster

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/config"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/logger"
)

// Manager handles Kind cluster operations
type Manager struct {
	cfg    *config.Config
	logger *logger.Logger
}

// NewManager creates a new cluster manager
func NewManager(ctx context.Context, cfg *config.Config, l *logger.Logger) (*Manager, error) {
	m := &Manager{
		cfg:    cfg,
		logger: l,
	}

	// Check if required tools are installed
	if err := m.checkRequiredTools(); err != nil {
		return nil, fmt.Errorf("failed to check required tools: %w", err)
	}

	// Create the cluster if it doesn't exist
	if !cfg.SkipClusterCreation {
		if err := m.createCluster(ctx); err != nil {
			return nil, fmt.Errorf("failed to create cluster: %w", err)
		}
	} else {
		// Check if the cluster exists
		exists, err := m.clusterExists(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check if cluster exists: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("cluster %s does not exist", cfg.ClusterName)
		}
		m.logger.Info("Using existing cluster: %s", cfg.ClusterName)
	}

	// Build and load the Docker image
	if !cfg.SkipBuild {
		if err := m.buildAndLoadImage(ctx); err != nil {
			return nil, fmt.Errorf("failed to build and load image: %w", err)
		}
	} else {
		m.logger.Info("Skipping Docker image build")
	}

	// Deploy the application
	if err := m.deployApplication(ctx); err != nil {
		return nil, fmt.Errorf("failed to deploy application: %w", err)
	}

	return m, nil
}

// Cleanup cleans up resources
func (m *Manager) Cleanup(ctx context.Context) error {
	m.logger.Info("Cleaning up resources")

	// Delete the Helm release
	m.logger.Debug("Deleting Helm release: %s", m.cfg.ReleaseName)
	cmd := exec.CommandContext(ctx, "helm", "uninstall", m.cfg.ReleaseName, "--namespace", m.cfg.Namespace)
	if err := m.runCommand(cmd); err != nil {
		m.logger.Error("Failed to delete Helm release: %v", err)
	}

	// Delete the cluster
	m.logger.Debug("Deleting Kind cluster: %s", m.cfg.ClusterName)
	cmd = exec.CommandContext(ctx, "kind", "delete", "cluster", "--name", m.cfg.ClusterName)
	if err := m.runCommand(cmd); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	return nil
}

// GetKubeconfig returns the kubeconfig path
func (m *Manager) GetKubeconfig() string {
	// Kind uses the default kubeconfig location
	home, err := os.UserHomeDir()
	if err != nil {
		m.logger.Error("Failed to get user home directory: %v", err)
		return ""
	}
	return fmt.Sprintf("%s/.kube/config", home)
}

// GetServiceEndpoint returns the endpoint for a service
func (m *Manager) GetServiceEndpoint(ctx context.Context, serviceName, portName string) (string, error) {
	// For Kind, we need to port-forward to access services
	// Find a free port
	freePort, err := m.findFreePort()
	if err != nil {
		return "", fmt.Errorf("failed to find free port: %w", err)
	}

	// Start port-forwarding in the background
	cmd := exec.CommandContext(ctx, "kubectl", "port-forward", 
		fmt.Sprintf("svc/%s", serviceName), 
		fmt.Sprintf("%d:%s", freePort, portName),
		"--namespace", m.cfg.Namespace)
	
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start port-forwarding: %w", err)
	}

	// Give it a moment to establish the connection
	time.Sleep(2 * time.Second)

	// Return the local endpoint
	return fmt.Sprintf("localhost:%d", freePort), nil
}

// GetGRPCEndpoint returns the gRPC endpoint
func (m *Manager) GetGRPCEndpoint(ctx context.Context) (string, error) {
	return m.GetServiceEndpoint(ctx, m.cfg.ReleaseName, fmt.Sprintf("%d", m.cfg.GRPCPort))
}

// ExecuteInPod executes a command in a pod
func (m *Manager) ExecuteInPod(ctx context.Context, podName, containerName, command string) (string, error) {
	args := []string{"exec", "-n", m.cfg.Namespace, podName}
	if containerName != "" {
		args = append(args, "-c", containerName)
	}
	args = append(args, "--", "sh", "-c", command)
	
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute command in pod: %w, stderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// WaitForDeployment waits for a deployment to be ready
func (m *Manager) WaitForDeployment(ctx context.Context, deploymentName string, timeout time.Duration) error {
	m.logger.Debug("Waiting for deployment %s to be ready", deploymentName)
	
	cmd := exec.CommandContext(ctx, "kubectl", "rollout", "status", 
		fmt.Sprintf("deployment/%s", deploymentName),
		"--namespace", m.cfg.Namespace,
		fmt.Sprintf("--timeout=%s", timeout))
	
	return m.runCommand(cmd)
}

// WaitForPod waits for a pod to be ready
func (m *Manager) WaitForPod(ctx context.Context, labelSelector string, timeout time.Duration) (string, error) {
	m.logger.Debug("Waiting for pod with selector %s to be ready", labelSelector)
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.CommandContext(ctx, "kubectl", "get", "pods",
			"--namespace", m.cfg.Namespace,
			"--selector", labelSelector,
			"--output", "jsonpath={.items[0].metadata.name}")
		
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		if err := cmd.Run(); err != nil {
			m.logger.Debug("Failed to get pod name: %v, stderr: %s", err, stderr.String())
			time.Sleep(2 * time.Second)
			continue
		}
		
		podName := strings.TrimSpace(stdout.String())
		if podName == "" {
			m.logger.Debug("No pod found with selector %s", labelSelector)
			time.Sleep(2 * time.Second)
			continue
		}
		
		// Check if the pod is ready
		cmd = exec.CommandContext(ctx, "kubectl", "get", "pod", podName,
			"--namespace", m.cfg.Namespace,
			"--output", "jsonpath={.status.phase}")
		
		stdout.Reset()
		stderr.Reset()
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		if err := cmd.Run(); err != nil {
			m.logger.Debug("Failed to get pod status: %v, stderr: %s", err, stderr.String())
			time.Sleep(2 * time.Second)
			continue
		}
		
		phase := strings.TrimSpace(stdout.String())
		if phase == "Running" {
			return podName, nil
		}
		
		m.logger.Debug("Pod %s is in phase %s, waiting...", podName, phase)
		time.Sleep(2 * time.Second)
	}
	
	return "", fmt.Errorf("timed out waiting for pod with selector %s", labelSelector)
}

// GetPodLogs gets logs from a pod
func (m *Manager) GetPodLogs(ctx context.Context, podName, containerName string) (string, error) {
	args := []string{"logs", "-n", m.cfg.Namespace, podName}
	if containerName != "" {
		args = append(args, "-c", containerName)
	}
	
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w, stderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// checkRequiredTools checks if required tools are installed
func (m *Manager) checkRequiredTools() error {
	tools := []string{"kind", "docker", "kubectl", "helm"}
	
	for _, tool := range tools {
		m.logger.Debug("Checking if %s is installed", tool)
		
		cmd := exec.Command("which", tool)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("required tool %s is not installed", tool)
		}
	}
	
	return nil
}

// clusterExists checks if the cluster exists
func (m *Manager) clusterExists(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to get clusters: %w", err)
	}
	
	clusters := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	for _, cluster := range clusters {
		if cluster == m.cfg.ClusterName {
			return true, nil
		}
	}
	
	return false, nil
}

// createCluster creates a Kind cluster
func (m *Manager) createCluster(ctx context.Context) error {
	// Check if the cluster already exists
	exists, err := m.clusterExists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}
	
	if exists {
		m.logger.Info("Cluster %s already exists", m.cfg.ClusterName)
		return nil
	}
	
	m.logger.Info("Creating Kind cluster: %s", m.cfg.ClusterName)
	
	// Create a temporary config file
	configFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(configFile.Name())
	
	// Write the config
	config := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8080
    hostPort: 8080
    protocol: TCP
  - containerPort: 50051
    hostPort: 50051
    protocol: TCP
`
	
	if _, err := configFile.WriteString(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	if err := configFile.Close(); err != nil {
		return fmt.Errorf("failed to close config file: %w", err)
	}
	
	// Create the cluster
	cmd := exec.CommandContext(ctx, "kind", "create", "cluster",
		"--name", m.cfg.ClusterName,
		"--config", configFile.Name())
	
	if err := m.runCommand(cmd); err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}
	
	return nil
}

// buildAndLoadImage builds and loads the Docker image
func (m *Manager) buildAndLoadImage(ctx context.Context) error {
	m.logger.Info("Building Docker image: %s:%s", m.cfg.ImageName, m.cfg.ImageTag)
	
	// Build the image
	cmd := exec.CommandContext(ctx, "docker", "build",
		"--build-arg", fmt.Sprintf("VERSION=%s", m.cfg.ImageTag),
		"-t", fmt.Sprintf("%s:%s", m.cfg.ImageName, m.cfg.ImageTag),
		".")
	
	if err := m.runCommand(cmd); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	
	// Load the image into Kind
	m.logger.Info("Loading image into Kind cluster")
	cmd = exec.CommandContext(ctx, "kind", "load", "docker-image",
		fmt.Sprintf("%s:%s", m.cfg.ImageName, m.cfg.ImageTag),
		"--name", m.cfg.ClusterName)
	
	if err := m.runCommand(cmd); err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	
	return nil
}

// deployApplication deploys the application
func (m *Manager) deployApplication(ctx context.Context) error {
	m.logger.Info("Deploying application")
	
	// Create a temporary values file
	valuesFile, err := os.CreateTemp("", "values-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(valuesFile.Name())
	
	// Write the values
	values := fmt.Sprintf(`image:
  repository: %s
  tag: %s
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  enabled: true
  ports:
    - name: http
      port: 8080
      targetPort: 8080
    - name: grpc
      port: %d
      targetPort: %d

worker:
  enabled: true
  replicas: 1
  concurrency: %d
  depth: %d
  fastMode: %t
  emailExtraction: false
  exitOnInactivity: "1h"
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi

config:
  grpc:
    enabled: true
    port: %d
    serviceName: "vector-leads-scraper"
    environment: "development"

  database:
    dsn: "postgres://postgres:postgres@%s-postgresql:5432/leads_scraper?sslmode=disable"
    maxIdleConnections: 10
    maxOpenConnections: 100
    maxConnectionLifetime: "10m"
    maxConnectionRetryTimeout: "10s"
    retrySleep: "1s"
    queryTimeout: "10s"
    maxConnectionRetries: 3

  redis:
    enabled: true
    host: "%s-redis-master"
    port: %d
    password: "%s"
    dsn: "redis://:redispass@%s-redis-master:%d/0"
    workers: 10
    retryInterval: "5s"
    maxRetries: 3
    retentionDays: 7

  newrelic:
    enabled: true
    key: "2aa111a8b39e0ebe981c11a11cc8792cFFFFNRAL"

  logging:
    level: "info"

  scraper:
    webServer: true
    concurrency: %d
    depth: %d
    language: "en"
    searchRadius: 10000
    zoomLevel: 15
    fastMode: %t

tests:
  enabled: true
  healthCheck:
    enabled: true
    path: "/health"
  configCheck:
    enabled: true

postgresql:
  enabled: true
  auth:
    username: "%s"
    password: "%s"
    database: "%s"
    existingSecret: ""
  primary:
    persistence:
      enabled: true
      size: 10Gi
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    service:
      ports:
        postgresql: %d

redis:
  enabled: true
  architecture: standalone
  auth:
    enabled: true
    password: "%s"
  master:
    persistence:
      enabled: true
      size: 1Gi
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 256Mi
`,
		m.cfg.ImageName,
		m.cfg.ImageTag,
		m.cfg.GRPCPort,
		m.cfg.GRPCPort,
		m.cfg.WorkerConcurrency,
		m.cfg.WorkerDepth,
		m.cfg.WorkerFastMode,
		m.cfg.GRPCPort,
		m.cfg.ReleaseName,
		m.cfg.ReleaseName,
		m.cfg.RedisPort,
		m.cfg.RedisPassword,
		m.cfg.ReleaseName,
		m.cfg.RedisPort,
		m.cfg.WorkerConcurrency,
		m.cfg.WorkerDepth,
		m.cfg.WorkerFastMode,
		m.cfg.PostgresUser,
		m.cfg.PostgresPassword,
		m.cfg.PostgresDatabase,
		m.cfg.PostgresPort,
		m.cfg.RedisPassword)
	
	if _, err := valuesFile.WriteString(values); err != nil {
		return fmt.Errorf("failed to write values: %w", err)
	}
	
	if err := valuesFile.Close(); err != nil {
		return fmt.Errorf("failed to close values file: %w", err)
	}
	
	// Deploy with Helm
	cmd := exec.CommandContext(ctx, "helm", "upgrade", "--install",
		m.cfg.ReleaseName,
		m.cfg.ChartPath,
		"--namespace", m.cfg.Namespace,
		"--create-namespace",
		"--values", valuesFile.Name(),
		"--values", "values.integration-tests.yaml")
	
	if err := m.runCommand(cmd); err != nil {
		return fmt.Errorf("failed to deploy application: %w", err)
	}
	
	// Wait for the deployment to be ready
	m.logger.Info("Waiting for deployment to be ready")
	if err := m.WaitForDeployment(ctx, m.cfg.ReleaseName, 5*time.Minute); err != nil {
		return fmt.Errorf("failed to wait for deployment: %w", err)
	}
	
	// Wait for PostgreSQL to be ready
	m.logger.Info("Waiting for PostgreSQL to be ready")
	if err := m.WaitForDeployment(ctx, fmt.Sprintf("%s-postgresql", m.cfg.ReleaseName), 5*time.Minute); err != nil {
		return fmt.Errorf("failed to wait for PostgreSQL: %w", err)
	}
	
	// Wait for Redis to be ready
	m.logger.Info("Waiting for Redis to be ready")
	if err := m.WaitForDeployment(ctx, fmt.Sprintf("%s-redis-master", m.cfg.ReleaseName), 5*time.Minute); err != nil {
		return fmt.Errorf("failed to wait for Redis: %w", err)
	}
	
	m.logger.Success("Application deployed successfully")
	
	return nil
}

// runCommand runs a command and logs its output
func (m *Manager) runCommand(cmd *exec.Cmd) error {
	m.logger.Debug("Running command: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		m.logger.Debug("Command failed: %v", err)
		m.logger.Debug("Stdout: %s", stdout.String())
		m.logger.Debug("Stderr: %s", stderr.String())
		return fmt.Errorf("command failed: %w", err)
	}
	
	m.logger.Debug("Command succeeded")
	m.logger.Debug("Stdout: %s", stdout.String())
	
	return nil
}

// findFreePort finds a free port
func (m *Manager) findFreePort() (int, error) {
	// For simplicity, we'll use a fixed port for now
	// In a real implementation, you would use net.Listen to find a free port
	return 8081, nil
}

// GetConfig returns the configuration
func (m *Manager) GetConfig() *config.Config {
	return m.cfg
} 