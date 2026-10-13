# Insecure Proxy

A lightweight, minimal dependency reverse proxy written in Go. It is designed to bridge HTTP/HTTPS clients to HTTPS upstream servers, specifically ignoring upstream certificate errors (insecure transport). It also supports automatic self-signed certificate generation for its own HTTPS listener.

## Features

- **HTTP Listener:** Proxies plain HTTP requests to HTTPS targets.
- **HTTPS Listener:** Proxies HTTPS requests to HTTPS targets.
- **Auto-Cert:** Automatically generates valid self-signed certificates if no files are provided.
- **Insecure Transport:** Ignores SSL certificate errors from upstream servers (useful for localhost/internal testing).
- **Endpoint Discovery:** `GET /endpoints` lists every route configured on the proxy.
- **Static binary:** Can be built as a single binary with no external dependencies.

## Build

To build the binary for your platform (standard Linux/macOS):

```bash
go build -o proxy cmd/proxy/main.go
```

For **RHEL 8/9** (or any Linux) efficiently (static link):

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o proxy cmd/proxy/main.go
```

## Configuration

The proxy is controlled by a YAML configuration file.

### Options

| Field | Description | Default |
|-------|-------------|---------|
| `server.addr` | The address to listen on (e.g., `:8080`) | **Required** |
| `server.tls` | Enable HTTPS listener | `false` |
| `server.cert_file` | Path to PEM encoded certificate file (optional) | Empty (Auto-gen) |
| `server.key_file` | Path to PEM encoded key file (optional) | Empty (Auto-gen) |
| `server.endpoints_path` | Path that lists all configured routes (optional) | `/endpoints` |
| `routes` | List of path mappings | **Required** |
| `routes[].path` | The path prefix to match (e.g., `/metrics`) | |
| `routes[].target` | The upstream URL to forward to (e.g., `https://localhost:9100/metrics`) | |

## Examples

### 1. HTTP Listener -> HTTPS Target

**Scenario:** You have a local exporter running on HTTPS with a bad/self-signed cert, and you want to scrape it via plain HTTP.

`config_http.yaml`:
```yaml
server:
  addr: ":8080"
  tls: false

routes:
  - path: "/metrics"
    target: "https://localhost:9100/metrics"
```

**Run:**
```bash
./proxy -config config_http.yaml
```

**Test:**
```bash
curl http://localhost:8080/metrics
```

---

### 2. HTTPS Listener (Auto Cert) -> HTTPS Target

**Scenario:** You want the proxy itself to be HTTPS, but you don't want to deal with generating certificates manually.

`config_https.yaml`:
```yaml
server:
  addr: ":8443"
  tls: true
  # No cert_file/key_file provided -> Auto-generates self-signed cert

routes:
  - path: "/metrics"
    target: "https://localhost:9100/metrics"
```

**Run:**
```bash
./proxy -config config_https.yaml
```

**Test:**
(Use `-k` to accept the auto-generated proxy cert)
```bash
curl -k -v https://localhost:8443/metrics
```

---

### 3. HTTPS Listener (Custom Cert) -> HTTPS Target

**Scenario:** You have your own certificates for the proxy.

`config_custom.yaml`:
```yaml
server:
  addr: ":8443"
  tls: true
  cert_file: "/path/to/server.crt"
  key_file: "/path/to/server.key"

routes:
  - path: "/app"
    target: "https://internal-app:443"
```

**Run:**
```bash
./proxy -config config_custom.yaml
```

---

### 4. Custom Interface Binding (Advanced)

**Scenario:** Your app is listening on a specific network interface (e.g., a private LAN IP `10.0.0.5`), not localhost.

`config_lan.yaml`:
```yaml
server:
  addr: "10.0.0.5:8443"     # Bind proxy to specific interface
  tls: true

routes:
  - path: "/"
    target: "https://10.0.0.5:9443" # Proxy to app on specific interface
```

**Run:**
```bash
./proxy -config config_lan.yaml
```

---

---

### 5. Custom Headers (Authorization etc.)

**Scenario:** The upstream metrics endpoint requires specific headers, such as an `Authorization` token.

`config_headers.yaml`:
```yaml
server:
  addr: ":8443"
  tls: true

routes:
  - path: "/secure-metrics"
    target: "https://secure-upstream:9100"
    headers:
      Authorization: "Basic YWRtaW46c2VjcmV0" # injected into upstream request
      X-Custom-Header: "foobar"
```

The proxy will automatically inject these headers into requests sent to the upstream.

---

### 6. Listing Configured Endpoints

**Scenario:** Each server runs the proxy with a different set of routes, and you want to see which endpoints a given server exposes.

`config_endpoints.yaml`:
```yaml
server:
  addr: ":8443"
  tls: true
  # endpoints_path: "/endpoints"  # Optional, this is the default

routes:
  - path: "/node"
    target: "https://localhost:9100/metrics"
  - path: "/opa"
    target: "https://localhost:8181/metrics"
```

**Test:**
```bash
curl -k https://localhost:8443/endpoints
```

```json
{"endpoints":["/node","/opa"]}
```

Every route in the config is listed, in config order, however many there are. Only paths are returned; targets and headers are never exposed. Routes whose target URL fails to parse are skipped, matching what the proxy actually serves. If a route is itself configured on `/endpoints`, that route wins and the list is not served; set `endpoints_path` to move it.

---

### 7. Running on Linux (RHEL 8/9)

To run on Linux, especially minimal environments like RHEL 8/9 or scratch containers, build a static binary:

```bash
./build_linux.sh
```

This produces `proxy-linux-amd64` which is a self-contained static binary.

