# Local model monorepo template

A deliberately small starting point for an application that uses a local GGUF model. It provides an OpenAI-compatible HTTP gateway and hardware-specific llama.cpp deployments, leaving the root free for your own `frontend/`, `backend/`, or other services.

## Layout

```text
.
├── local-ai/
│   ├── Dockerfile
│   ├── go.mod
│   └── server/main.go
├── deploy/
│   ├── cpu/docker-compose.yml
│   ├── nvidia/docker-compose.yml
│   ├── rocm/docker-compose.yml
│   └── intel/docker-compose.yml
├── models/                  # downloaded GGUF files; ignored by Git
├── scripts/download-model.sh
└── docker-compose.yml       # add your application services here
```

## Start it

Docker Compose v2 and `curl` are required. Download the default model:

```sh
./scripts/download-model.sh
```

Start the base stack with the override matching your hardware:

```sh
# CPU
docker compose -f docker-compose.yml -f deploy/cpu/docker-compose.yml up --build

# NVIDIA CUDA
docker compose -f docker-compose.yml -f deploy/nvidia/docker-compose.yml up --build

# AMD ROCm
docker compose -f docker-compose.yml -f deploy/rocm/docker-compose.yml up --build

# Intel SYCL
docker compose -f docker-compose.yml -f deploy/intel/docker-compose.yml up --build
```

The gateway is bound to `127.0.0.1:8080`. Startup is complete when this returns `{"status":"ok"}`:

```sh
curl http://127.0.0.1:8080/healthz
```

Use it with an OpenAI-compatible client by setting its base URL to `http://127.0.0.1:8080/v1`. No API key is required locally. For example:

```sh
curl http://127.0.0.1:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"local","messages":[{"role":"user","content":"Hello!"}]}'
```

## Configure it

| Variable | Default | Purpose |
| --- | --- | --- |
| `LOCAL_AI_PORT` | `8080` | Gateway port on the host |
| `MODEL_FILE` | `smollm2-1.7b-instruct-q4_k_m.gguf` | File inside `models/` |
| `CONTEXT_SIZE` | `8192` | Model context size |
| `PARALLEL_REQUESTS` | `1` | Concurrent llama.cpp slots |
| `GPU_LAYERS` | `99` | Layers offloaded in GPU deployments |

The download script also supports `135m` and `360m` presets. If you select one, set `MODEL_FILE` when starting Compose.

To add an application, define its service in the root `docker-compose.yml`. Services that need private model access can join the `inference` network and call `http://local-ai:8080/v1`; host applications can use `http://127.0.0.1:8080/v1`.

The gateway has no authentication and binds to loopback intentionally. Add authentication and TLS before exposing it beyond the local machine.
