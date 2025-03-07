# Use Ubuntu as the base image
FROM ubuntu:latest

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Add build arguments for user and group IDs
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG GIT_USER_EMAIL
ARG GIT_USER_NAME

# Update package lists and upgrade existing packages
RUN apt-get update && apt-get upgrade -y

# Install required packages
RUN apt-get install -y \
    git \
    curl \
    build-essential \
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    libxml2-dev \
    libxmlsec1-dev \
    libffi-dev \
    liblzma-dev \
    iputils-ping \
    htop \
    zsh \
    wget \
    python3 \
    python3-full \
    python3-venv \
    ffmpeg \
    vim \
    jq

# Handle user/group creation
RUN set -eux; \
    USER_ID="${USER_ID:-1000}"; \
    GROUP_ID="${GROUP_ID:-1000}"; \
    # Create or modify group to match target GID
    if getent group "${GROUP_ID}" >/dev/null; then \
        existing_group_name=$(getent group "${GROUP_ID}" | cut -d: -f1); \
        if [ "$existing_group_name" != "appuser" ]; then \
            groupmod -n appuser "$existing_group_name"; \
        fi; \
    else \
        groupadd -g "${GROUP_ID}" appuser; \
    fi; \
    # Handle user creation
    if getent passwd "${USER_ID}" >/dev/null; then \
        existing_user_name=$(getent passwd "${USER_ID}" | cut -d: -f1); \
        if [ "$existing_user_name" != "appuser" ]; then \
            userdel -f "$existing_user_name"; \
        fi; \
    fi; \
    useradd -u "${USER_ID}" -g appuser -G appuser -m -d /home/appuser appuser; \
    # Ensure group membership is correct
    usermod -a -G appuser appuser

# Create virtual environment directory and set ownership
RUN mkdir -p /opt/venv && chown appuser:appuser /opt/venv

# Switch to appuser
USER appuser

# Set environment variables for the new user
ENV HOME=/home/appuser
ENV PATH="${HOME}/.local/bin:${PATH}"

# Set up virtual environment
ENV VIRTUAL_ENV=/opt/venv
RUN python3 -m venv $VIRTUAL_ENV
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

# Git configuration
RUN git config --global user.email "${GIT_USER_EMAIL}" && \
    git config --global user.name "${GIT_USER_NAME}" && \
    git config --global --add safe.directory '*'

# Install oh-my-zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

# Install pyenv
RUN curl https://pyenv.run | bash

# Set up pyenv environment variables
ENV PYENV_ROOT=$HOME/.pyenv
ENV PATH=$PYENV_ROOT/bin:$PATH

# Initialize pyenv and virtual environment in zshrc
RUN echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.zshrc && \
    echo 'export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.zshrc && \
    echo 'eval "$(pyenv init --path)"' >> ~/.zshrc && \
    echo 'eval "$(pyenv init -)"' >> ~/.zshrc && \
    echo 'export VIRTUAL_ENV=/opt/venv' >> ~/.zshrc && \
    echo 'export PATH="$VIRTUAL_ENV/bin:$PATH"' >> ~/.zshrc

# pip install with dependencies
RUN pip install --upgrade pip && \
    pip install python-redmine && \
    pip install aider-install && \
    aider-install && \
    pip install aider && \
    pip install mirascope[all] && \
    pip install matplotlib

# Set zsh as default shell (requires root)
USER root
RUN chsh -s $(which zsh) appuser \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy executables and configuration files
COPY --chown=appuser:appuser build/ /usr/local/bin/
COPY --chown=appuser:appuser .andai.*.yaml /app/

# Set working directory
WORKDIR /app

# Final runtime configuration
USER appuser
SHELL ["/bin/zsh", "-c"]
ENTRYPOINT ["andai"]
