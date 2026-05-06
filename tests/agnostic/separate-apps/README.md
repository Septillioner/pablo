# Separate Apps Test Scenario

This directory contains **separate applications**, each with its own `pablo.yaml` file.

## Structure

```
separate-apps/
├── frontend/
│   └── pablo.yaml          (React/Vue/Angular)
├── backend/
│   └── pablo.yaml          (Docker Backend)
├── go-binary/
│   └── pablo.yaml          (Go Microservice)
└── php-app/
    └── pablo.yaml          (Laravel/PHP)
```

## Usage

Each app is deployed independently:

```bash
# Deploy frontend
cd frontend
pablo run -e production

# Deploy backend
cd ../backend
pablo run -e staging

# Deploy Go service
cd ../go-binary
pablo run -e development

# Deploy PHP app
cd ../php-app
pablo run -e production
```

## Apps

### 1. Frontend (`frontend/`)
- **Type:** `static`
- **Tech:** React/Vue/Angular
- **Environments:** development, production

### 2. Backend (`backend/`)
- **Type:** `docker`
- **Tech:** Node.js/Python API
- **Environments:** staging, production

### 3. Go Binary (`go-binary/`)
- **Type:** `binary`
- **Tech:** Go microservice
- **Environments:** development, production

### 4. PHP App (`php-app/`)
- **Type:** `git-sync`
- **Tech:** Laravel
- **Environments:** staging, production

## When to Use This Approach

✅ **Use separate apps when:**
- Each app has its own repository
- Different teams manage different apps
- Apps are deployed independently
- You want simpler, focused configs

❌ **Use multi-profile when:**
- All apps are in a monorepo
- You want centralized credential management
- Apps are deployed together
- You need to share common settings
