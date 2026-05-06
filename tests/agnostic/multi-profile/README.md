# Multi-Profile Test Scenario

This directory contains a **single `pablo.yaml`** file with multiple application profiles.

## Usage

```bash
# Deploy frontend to production
pablo run -p frontend -e production -f pablo.yaml

# Deploy backend to staging
pablo run -p backend-api -e staging -f pablo.yaml

# Deploy Go service to development
pablo run -p go-service -e development -f pablo.yaml

# Deploy PHP app to production
pablo run -p php-webapp -e production -f pablo.yaml
```

## Profiles

1. **frontend** (type: `static`)
   - React/Vue/Angular SPA
   - Build → Deploy static files

2. **backend-api** (type: `docker`)
   - Containerized backend service
   - Git pull → Docker compose up

3. **go-service** (type: `binary`)
   - Compiled Go binary
   - Build → Deploy → Systemd restart

4. **php-webapp** (type: `git-sync`)
   - PHP Laravel/Symfony app
   - Git pull → Composer install → Artisan commands

## Credentials

All profiles share global credentials:
- `prod_server`: SSH key for production
- `staging_server`: SSH password for staging
- `github_token`: GitHub access token
- `docker_registry`: Docker registry credentials
