# PlatformPilot

PlatformPilot is a lightweight Internal Developer Platform (IDP) built with Go and Kubernetes.

The idea is simple: developers should be able to deploy and manage applications without needing to know every Kubernetes command behind the scenes.

Instead of asking developers to work directly with `kubectl`, PlatformPilot provides a simple interface that handles common Kubernetes operations through a Go backend.

## Architecture

![PlatformPilot Architecture](docs/images/architecture%20diagram.png)

The current request flow is:

**Developer → PlatformPilot UI → Go REST API → Kubernetes client-go → Kubernetes API Server → Kubernetes Resources**

PlatformPilot currently runs against a Kubernetes cluster and manages resources inside the `platformpilot` namespace.

---

## What it can do

### Application deployment

Applications can be deployed through the PlatformPilot UI instead of manually creating Kubernetes manifests.

For example, a developer can provide:

- Application name
- Container image
- Number of replicas

PlatformPilot creates the required Kubernetes resources.

### Resource explorer

The UI provides a simple view of Kubernetes resources managed by PlatformPilot.

Currently supported:

- Deployments
- Pods
- Services

The dashboard shows:

- Resource name
- Resource type
- Status
- Replica information
- Available actions

### Resource deletion

Resources can be deleted directly from the UI.

PlatformPilot also protects its own critical resources from accidental deletion.

### Application status

PlatformPilot checks Kubernetes Deployment status and shows whether an application is running or still progressing.

### Scaling

Applications can be scaled by changing the desired replica count.

### Kubernetes self-healing

PlatformPilot relies on Kubernetes controllers for workload recovery.

For example, if a Pod belonging to a Deployment is deleted, Kubernetes automatically creates a replacement Pod.

---

## Technology Stack

- **Go** — Backend API
- **Kubernetes** — Container orchestration
- **client-go** — Kubernetes API integration
- **HTML / CSS / JavaScript** — Frontend
- **Docker** — Containerization
- **Kind** — Local Kubernetes development
- **GitHub** — Source control

---

## Project Structure

```text
platformpilot/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── api/
│   │   ├── deploy.go
│   │   ├── delete.go
│   │   ├── delete_resource.go
│   │   ├── resources.go
│   │   ├── scale.go
│   │   ├── status.go
│   │   └── ui.go
│   │
│   ├── kubernetes/
│   │   └── client.go
│   │
│   └── platform/
│       ├── deployment.go
│       ├── service.go
│       ├── scale.go
│       ├── status.go
│       ├── delete.go
│       ├── delete_resource.go
│       └── resources.go
│
├── web/
│   ├── templates/
│   │   └── index.html
│   │
│   └── static/
│       ├── app.js
│       └── style.css
│
├── docs/
│   └── images/
│       └── architecture diagram.png
│
├── Dockerfile
├── go.mod
└── README.md
