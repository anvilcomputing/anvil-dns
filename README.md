# Anvil DNS Manager

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

A lightweight, Go-based application for managing Cloudflare DNS records for `anvilcomputing.com`. 

Built with a modular architecture, this tool provides both a **Command Line Interface (CLI)** and a **Web Admin Dashboard**. It is designed to quickly provision unproxied `A` records that point to a VPS running a reverse proxy (like Pangolin), prevent naming collisions, and provide simple commands to audit and clean up your Cloudflare zone.

## ✨ Features

*   **Dual Interfaces:** Manage DNS via the terminal or a data-dense, filtering-enabled web dashboard.
*   **Collision Prevention:** Automatically checks if a subdomain exists before provisioning to prevent accidental overwrites.
*   **Smart Configuration:** Remembers your last used Target IP.
*   **Interactive Auditing:** Built-in paginated `list` command (CLI) and instant client-side filtering (Web) to easily audit existing records.
*   **Safe Deletion:** Includes confirmation prompts before deleting routing records to prevent accidents.

## 🚀 Prerequisites

1.  **Go 1.21+** installed on your machine.
2.  A **Cloudflare API Token** with the following permissions:
    *   `Zone` -> `DNS` -> `Edit`
    *   *Zone Resources:* `Include` -> `Specific Zone` -> `anvilcomputing.com`

## ⚙️ Configuration

Export your Cloudflare API token into your shell environment:
```bash
export CLOUDFLARE_API_TOKEN="your_custom_cloudflare_token"
```

*Optional:* Set a default Target IP for provisioning:
```bash
export TARGET_IP="<VPS_IP>"
```

---

## 💻 Usage: Command Line Interface (CLI)

To build or run the CLI:

```bash
# Check availability
go run cmd/cli/main.go check <username>

# Provision a new record
go run cmd/cli/main.go create <username> --target-ip <VPS_IP>

# Audit records (Paginated)
go run cmd/cli/main.go list

# Delete a record
go run cmd/cli/main.go delete <username>
```

---

## 🌐 Usage: Web Admin Dashboard

The Web Admin provides a responsive, data-dense UI to perform the same actions as the CLI. It runs a lightweight HTTP server on port `8081`.

```bash
go run cmd/web/main.go
```
Once running, open your browser and navigate to: **`http://localhost:8081`**

**Web Features:**
*   Create records using a simple form.
*   Instantly filter active subdomains by Name or Record Type without page reloads.
*   Delete records with one click (includes safety confirmations).
*   No horizontal scrolling; optimized for desktop data density.

---

## 📂 Project Structure

```text
anvil-dns/
├── cmd/
│   ├── cli/              # The CLI interface (Cobra/Viper)
│   └── web/              # The Web Admin interface (net/http, html/template)
│       ├── main.go
│       └── index.html
├── internal/
│   └── cloudflare/       # Shared Core API logic
├── go.mod
└── go.sum
```

## 🗺️ Future Enhancements

- [x] **`web` Interface (Admin UI):** Create a lightweight, data-dense HTML frontend.
- [ ] **NixOS & Proxmox Packaging:** 
  - Add a `flake.nix` to package the Go binaries.
  - Create a declarative NixOS LXC configuration for Proxmox.
  - Implement `sops-nix` for secure, encrypted Cloudflare API token management.
  - Enable continuous deployment via `deploy-rs`.
- [ ] **`api` Interface:** Add `cmd/api/main.go` to expose the core logic as a REST/Webhook interface for CI/CD automation agents.
- [ ] **Cloudflare Proxy Support:** Add a flag/toggle to optionally enable the Cloudflare proxy ("orange cloud").
