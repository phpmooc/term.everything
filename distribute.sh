#!/bin/bash

# This script builds a distributable AppImage
# of the term.everything application using Podman.

PODMAN="podman "
APP_NAME="term.everything❗mmulet.com-dont_forget_to_chmod_+x_this_file"


get_distro() {
    # Try to detect the distro
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
    else
        DISTRO="unknown"
    fi
    
    case $DISTRO in
        ubuntu|debian)
            echo "sudo apt update && sudo apt install -y "
            ;;
        fedora)
            echo "sudo dnf install -y "
            ;;
        centos|rhel|rocky|almalinux)
            echo "sudo yum install -y "
            ;;
        arch|manjaro)
            echo "sudo pacman -S "
            ;;
        opensuse*)
            echo "sudo zypper install "
            ;;
        alpine)
            echo "sudo apk add "
            ;;
        *)
            echo "Please install podman using your distribution's package manager"
            ;;
    esac
}

if ! command -v podman >/dev/null 2>&1; then
    if command -v docker >/dev/null 2>&1; then
        PODMAN="docker "
    else
        INSTALL_CMD=$(get_distro)
        echo "Warning: podman is not installed or not in PATH."
        echo "To install on your system, try: $INSTALL_CMD podman"
        echo "Please install podman to proceed, it's literally all you need. Don't even need attention. Just podman. Just get podman. What are you waiting for? Stop reading this and install podman."
        exit 1
    fi
  
fi

if [ -z "${PLATFORM+x}" ]; then
    PLATFORM_ARG=""
else
    PLATFORM_ARG="--platform $PLATFORM -e PLATFORM=$PLATFORM -e MULTI_PLATFORM=1 -e ARCH_PREFIX=$ARCH_PREFIX"
fi

$PODMAN run \
    $PLATFORM_ARG \
    -it \
    --volume .:/home/mount \
    --rm alpine:latest /bin/sh /home/mount/resources/alpineCompile.sh && \
echo "Output is ./dist/$PLATFORM/static/$APP_NAME"

