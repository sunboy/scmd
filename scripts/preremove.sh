#!/bin/sh
#
# Pre-remove script for scmd packages (deb, rpm, apk)
# Runs before package removal

set -e

echo "Removing scmd..."

# Note: The package manager will remove the binary and completions
# We don't remove ~/.scmd here as it contains user data

echo ""
echo "Note: User data in ~/.scmd has been preserved."
echo "To completely remove scmd data, run:"
echo "  rm -rf ~/.scmd"
echo ""

exit 0
