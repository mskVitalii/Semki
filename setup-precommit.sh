#!/bin/bash

HOOK_DIR=".git/hooks"
HOOK_FILE="$HOOK_DIR/pre-commit"

if [ ! -d "$HOOK_DIR" ]; then
    echo "This isn't Git repo"
    exit 1
fi

cat > "$HOOK_FILE" <<'EOF'
#!/bin/bash

cd ./server || exit 1

echo "Running go fmt..."
go fmt ./...
if [ $? -ne 0 ]; then
    echo "go fmt failed"
    exit 1
fi

echo "Running go vet..."
go vet ./...
if [ $? -ne 0 ]; then
    echo "go vet failed"
    exit 1
fi

echo "Generating swagger..."
make swagger
if [ $? -ne 0 ]; then
    echo "Swagger generation failed"
    exit 1
fi

echo "Pre-commit checks passed."
EOF

chmod +x "$HOOK_FILE"

echo "Pre-commit hook created!"
