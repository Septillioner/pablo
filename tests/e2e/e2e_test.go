//go:build integration

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("PABLO_E2E_SKIP_DOCKER") == "1" {
		os.Exit(m.Run())
	}

	if err := setupE2E(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	teardownE2E()
	os.Exit(code)
}

func TestSSH_StaticSite(t *testing.T) {
	dir := scenarioDir(t, "static-site")
	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "web", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "web", "-e", "production", "--force")

	assertRemoteFile(t, "/var/www/static-site/index.html")
	assertRemoteFile(t, "/var/www/static-site/assets/app.css")
	assertRemoteContains(t, "/var/www/static-site/index.html", "Acme Marketing")
}

func TestSSH_StaticSiteHotfix(t *testing.T) {
	dir := scenarioDir(t, "static-site-hotfix")
	target := "/var/www/static-site-hotfix"

	sshExec(t, fmt.Sprintf("rm -rf %s && mkdir -p %s/assets", target, target))
	sshExec(t, fmt.Sprintf("printf 'old-banner' > %s/index.html", target))
	sshExec(t, fmt.Sprintf("printf 'old' > %s/assets/app.css", target))

	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "web", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "web", "-e", "production", "--force")

	assertRemoteContains(t, target+"/index.html", "Hotfix deployed")
	content := strings.TrimSpace(sshExec(t, fmt.Sprintf("cat %s/index.html", target)))
	if strings.Contains(content, "old-banner") {
		t.Fatal("expected old seed content to be replaced")
	}
	listing := strings.TrimSpace(sshExec(t, fmt.Sprintf("ls -1 %s", target)))
	if strings.Contains(listing, "__renamed__") || strings.Contains(listing, ".bak") {
		t.Fatalf("unexpected leftover rename artifacts:\n%s", listing)
	}
	assertRemoteFile(t, target+"/assets/app.css")
}

func TestSSH_GoService(t *testing.T) {
	dir := scenarioDir(t, "go-service")
	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "default", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "default", "-e", "production", "--force")

	assertRemoteFile(t, "/opt/go-service/myservice")
	assertRemoteContains(t, "/opt/go-service/VERSION.txt", "1.2.0")
}

func TestSSH_ComposeAPI(t *testing.T) {
	dir := scenarioDir(t, "compose-api")
	sshExec(t, "docker rm -f pablo-e2e-sample 2>/dev/null || true")

	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "default", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "default", "-e", "production", "--force")
	assertComposeAPIRunning(t)

	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "default", "-e", "production", "--force")
	assertComposeAPIRunning(t)
}

func assertComposeAPIRunning(t *testing.T) {
	t.Helper()
	out := sshExec(t, "docker ps --format '{{.Names}}'")
	if !strings.Contains(out, "pablo-e2e-sample") {
		t.Fatalf("expected container pablo-e2e-sample in docker ps, got: %q", out)
	}
}

func TestSSH_PHPApp(t *testing.T) {
	dir := scenarioDir(t, "php-app")
	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "default", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "default", "-e", "production", "--force")

	assertRemoteFile(t, "/var/www/php-app/public/index.php")
	assertRemoteFile(t, "/var/www/php-app/storage/.deployed")
	assertRemoteFile(t, "/var/www/php-app/.env")
	assertRemoteContains(t, "/var/www/php-app/.env", "APP_ENV=production")
}

func TestSSH_ReleaseSequence(t *testing.T) {
	dir := scenarioDir(t, "release-sequence")
	runPablo(t, dir, "check", "-f", "pablo.yaml")
	runPablo(t, dir, "run", "sequence", "release", "-f", "pablo.yaml", "--force")

	assertRemoteFile(t, "/var/www/app-staging/index.html")
	assertRemoteFile(t, "/var/www/app/index.html")
	assertRemoteContains(t, "/var/www/app/index.html", "Release Web")
}

func TestSSH_SiteWithBackup(t *testing.T) {
	dir := scenarioDir(t, "site-with-backup")
	target := "/var/www/acme-blog"

	sshExec(t, fmt.Sprintf("rm -rf %s %s_backup_* && mkdir -p %s", target, target, target))
	sshExec(t, fmt.Sprintf("printf 'live-v1' > %s/index.html", target))

	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "web", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "web", "-e", "production", "--force")

	assertRemoteContains(t, target+"/index.html", "Acme Blog")
	backups := strings.TrimSpace(sshExec(t, fmt.Sprintf("ls -d %s_backup_* 2>/dev/null | head -n 1", target)))
	if backups == "" {
		t.Fatal("expected a backup directory sibling after strategy: backup")
	}
	assertRemoteContains(t, backups+"/index.html", "live-v1")
}

func TestSSH_CleanRedeploy(t *testing.T) {
	dir := scenarioDir(t, "clean-redeploy")
	target := "/var/www/docs-portal"

	sshExec(t, fmt.Sprintf("rm -rf %s && mkdir -p %s", target, target))
	sshExec(t, fmt.Sprintf("printf 'stale' > %s/stale.txt", target))
	sshExec(t, fmt.Sprintf("printf 'old' > %s/index.html", target))

	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "web", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "web", "-e", "production", "--force")

	assertRemoteContains(t, target+"/index.html", "Docs Portal")
	assertRemoteFile(t, target+"/assets/app.css")
	out := sshExec(t, fmt.Sprintf("test ! -f %s/stale.txt && echo gone", target))
	if !strings.Contains(out, "gone") {
		t.Fatal("expected stale.txt to be removed by strategy: recreate")
	}
}

func TestSSH_LegacyTransfer(t *testing.T) {
	dir := scenarioDir(t, "legacy-transfer")
	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "web", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "web", "-e", "production", "--force")

	assertRemoteFile(t, "/var/www/legacy-transfer/index.html")
	assertRemoteFile(t, "/var/www/legacy-transfer/assets/logo.svg")
	assertRemoteContains(t, "/var/www/legacy-transfer/index.html", "Legacy Transfer Site")
}

func TestSSH_VerifiedTransfer(t *testing.T) {
	dir := scenarioDir(t, "verified-transfer")
	runPablo(t, dir, "check", "-f", "pablo.yaml", "-p", "web", "-e", "production")
	runPablo(t, dir, "run", "-f", "pablo.yaml", "-p", "web", "-e", "production", "--force")

	assertRemoteFile(t, "/var/www/verified-site/index.html")
	assertRemoteContains(t, "/var/www/verified-site/index.html", "Verified Site")
}
