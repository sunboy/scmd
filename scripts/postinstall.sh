#!/bin/sh
#
# Post-install script for scmd packages (deb, rpm, apk)
# Runs after package installation

set -e

echo "scmd has been installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Verify installation: scmd --version"
echo "  2. For offline usage, install llama.cpp:"
echo "     - Debian/Ubuntu: apt-get install llama-cpp"
echo "     - Fedora/RHEL: dnf install llama-cpp"
echo "     - Or build from source: https://github.com/ggerganov/llama.cpp"
echo "  3. Try it out: scmd /explain \"what is a goroutine\""
echo ""
echo "Documentation: https://github.com/scmd/scmd"
echo ""

# Reload shell completions (if shell is running)
if [ -n "$BASH_VERSION" ]; then
    if [ -f /usr/share/bash-completion/completions/scmd ]; then
        echo "Bash completion installed. Restart your shell or run:"
        echo "  source /usr/share/bash-completion/completions/scmd"
    fi
fi

exit 0
