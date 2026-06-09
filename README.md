# **Anvil DNS Manager**

A lightweight, Go-based application for managing Cloudflare DNS records for `anvilcomputing.com`.

Currently operating as a CLI, this tool is designed to quickly provision unproxied `A` records that point to a VPS running a reverse proxy (like Pangolin). It ensures naming collisions are prevented and provides simple commands to audit and clean up your Cloudflare zone.

## **✨ Features**

* **Fast & Lightweight:** Compiled as a single static Go binary.  
* **Collision Prevention:** Automatically checks if a subdomain exists before provisioning to prevent accidental overwrites.  
* **Smart Configuration:** Remembers your last used Target IP using local configuration (`~/.anvil-dns.yaml`).  
* **Interactive Auditing:** Built-in paginated `list` command to easily audit all existing DNS records in the terminal.  
* **Safe Deletion:** Includes confirmation prompts before deleting routing records.

## **🚀 Prerequisites**

1. **Go 1.21+** installed on your machine.  
2. A **Cloudflare API Token** with the following permissions:  
   * `Zone` \-\> `DNS` \-\> `Edit`  
   * *Zone Resources:* `Include` \-\> `Specific Zone` \-\> `anvilcomputing.com`

## **⚙️ Configuration**

The application relies on environment variables and a local configuration file.

### **1\. Authentication**

Export your Cloudflare API token into your shell environment:

```shell
export CLOUDFLARE_API_TOKEN="your_custom_cloudflare_token"
```

### **2\. Target IP (For Provisioning)**

When creating a new record, the app needs to know where to point it. You can provide this in three ways (in order of precedence):

1. **CLI Flag:** `--target-ip <ip_of_VPS>`  
2. **Environment Variable:** `export TARGET_IP="<ip_of_VPS>"`  
3. **Local Config (Automatic):** Upon a successful creation, the app saves the IP to `~/.anvil-dns.yaml` and will default to it for future runs.

## **🛠️ Usage**

To build the application locally:

```shell
go build -o anvil-dns cmd/cli/main.go
```

### **Available Commands**

#### **1\. Check a Record**

Verify if a specific username/subdomain is available or already in use.

```shell
./anvil-dns check <username>
```

#### **2\. Create a Record**

Provision a new `A` record. *(Requires `--target-ip` on the first run).*

```shell
./anvil-dns create <username> --target-ip <ip_of_VPS>
```

#### **3\. List Records**

Interactively page through all DNS records in the `anvilcomputing.com` zone.

```shell
./anvil-dns list
```

#### **4\. Delete a Record**

Safely remove a DNS record. Prompts for confirmation before execution.

```shell
./anvil-dns delete <username>
```

## **📂 Project Structure**

```
anvil-dns/
├── cmd/
│   └── cli/              # The current CLI interface (Cobra/Viper)
├── internal/
│   └── cloudflare/       # Core API logic (separated for reuse in future Web/API interfaces)
├── go.mod
└── go.sum
```

## **🗺️ Future Enhancements**

This project is built with a modular architecture to support future expansion into a multi-interface service deployed on NixOS.

* **NixOS & Proxmox Packaging:**  
  * Add a `flake.nix` to package the Go binary.  
  * Create a declarative NixOS LXC configuration for Proxmox.  
  * Implement `sops-nix` for secure, encrypted Cloudflare API token management.  
  * Enable continuous deployment via `deploy-rs`.  
* **`api` Interface:**  
  * Add `cmd/api/main.go` to expose the core logic as a REST/Webhook interface for CI/CD automation agents.  
* **`web` Interface (Admin UI):**  
  * Add `cmd/web/main.go` to serve a lightweight, mobile-friendly HTML frontend. This will provide a human-usable graphical dashboard to list, create, and delete records from a phone or browser without needing terminal access.  
* **Cloudflare Proxy Support:** Add a flag/toggle to optionally enable the Cloudflare proxy ("orange cloud") once the Pangolin reverse proxy is configured to handle Cloudflare Origin certificates and real IP resolution.
