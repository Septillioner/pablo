# Pablo terminal demo — init → check → run (~30s)
# Render: vhs docs/assets/demo.tape
# Requires: vhs (https://github.com/charmbracelet/vhs) and `pablo` on PATH
#   inside the VHS environment (Linux container / host where vhs runs).

Output docs/assets/demo.gif

Require pablo

Set Shell bash
Set FontSize 16
Set Width 1100
Set Height 600
Set TypingSpeed 40ms
Set Theme "Catppuccin Mocha"

Hide
Type "mkdir -p /tmp/pablo-demo/src && cd /tmp/pablo-demo"
Enter
Type "printf '%s\n' '<!DOCTYPE html><html><body><h1>Hello from Pablo</h1></body></html>' > src/index.html"
Enter
Type "cat > pablo.yaml << 'EOF'"
Enter
Type "name: demo-app"
Enter
Type "version: 0.1.0"
Enter
Type "profiles:"
Enter
Type "  default:"
Enter
Type "    type: static"
Enter
Type "    environments:"
Enter
Type "      production:"
Enter
Type "        deploy:"
Enter
Type "          source:"
Enter
Type "            dir: ./src"
Enter
Type `            include: ["**/*"]`
Enter
Type "          target_path: ./deploy-output"
Enter
Type "          strategy: overwrite"
Enter
Type "EOF"
Enter
Show

Type "pablo init"
Enter
Sleep 1.2s

Type "ls pablo_sample.yaml pablo.yaml src"
Enter
Sleep 800ms

Type "pablo check -f pablo.yaml"
Enter
Sleep 1.2s

Type "pablo run -f pablo.yaml -p default -e production"
Enter
Sleep 2.5s

Type "ls deploy-output"
Enter
Sleep 1.5s
