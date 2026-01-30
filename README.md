# Gimpel

<img src="docs/assets/logo.png" alt="Gimpel Logo" width="200" />

Gimpel is a distributed system for managing security monitoring modules (honeypots, sensors, etc.) across multiple servers. It provides centralized control, telemetry collection, high-interactive sessions as well as remote deployment with a focus on scalability and extensibility.

Key components:

- **Master**: Central control plane for managing deployments, pairings, and module catalog
- **Agent**: Runs on target hosts to execute security modules and forward telemetry
- **Gateway**: Collects and ingests telemetry data from agents
- **Sandbox**: Isolated environment for high-interactive sessions
- **SDK**: Libraries for building custom security modules (currently only Go; Java, Python and Rust are planned)

## Disclaimer

This project is in an early stage of development and is not ready for production use.

The legal stuff:
Gimpel is provided on an “as is” and “as available” basis, without warranties of any kind, express or implied. The security properties of the system have not been validated and may change at any time; the Web UI is also still under active development.

You are responsible for evaluating the software, its configuration, and its suitability for your environment. Use at your own risk. To the maximum extent permitted by applicable law, the author(s) and contributor(s) are not liable for any direct, indirect, incidental, or consequential damages, or for any loss resulting from use or inability to use the software (including security incidents).

## Installation

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for containerized deployment)
- Protocol Buffers compiler (for development)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/Kamuyin/gimpel.git
cd gimpel

# Build all services
go build -o bin/master ./cmd/master
go build -o bin/agent ./cmd/agent
go build -o bin/gateway ./cmd/gateway
go build -o bin/sandbox ./cmd/sandbox
```

### Docker Deployment

```bash
cd deploy/docker
docker compose build
docker compose up -d
```

This starts all services with default configurations in `deploy/docker/config/`.
You will need to adjust the Docker configurations of the agents to expose the services from the modules.

### Running Individual Services

```bash
# Master (control plane)
./bin/master -config /etc/gimpel/master.yaml

# Agent (on target hosts)
./bin/agent -config /etc/gimpel/agent.yaml

# Gateway (telemetry collector)
./bin/gateway -config /etc/gimpel/gateway.yaml

# Sandbox (testing environment)
./bin/sandbox -config /etc/gimpel/sandbox.yaml
```

### Module Development

To be documented.

## Configuration

Each service requires a YAML configuration file. Examples are in `deploy/docker/config/`.

## License

TBD
