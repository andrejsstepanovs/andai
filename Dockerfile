# Use Ubuntu as the base image
FROM ubuntu:latest

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive
ENV PATH="/root/.local/bin:${PATH}"

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

# Create a virtual environment
ENV VIRTUAL_ENV=/opt/venv
RUN python3 -m venv $VIRTUAL_ENV
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

# Install oh-my-zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

# Install pyenv
RUN curl https://pyenv.run | bash

# Set up pyenv environment variables
ENV HOME=/root
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
RUN . $VIRTUAL_ENV/bin/activate && pip install --upgrade pip
RUN . $VIRTUAL_ENV/bin/activate && pip install python-redmine
RUN . $VIRTUAL_ENV/bin/activate && pip install aider-install && aider-install && pip install aider
RUN . $VIRTUAL_ENV/bin/activate && pip install mirascope[all]
RUN . $VIRTUAL_ENV/bin/activate && pip install matplotlib

RUN git config --global user.email "${GIT_USER_EMAIL}" && \
    git config --global user.name "${GIT_USER_NAME}"

# Set zsh as default shell
RUN chsh -s $(which zsh)

# Clean up
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# copy executables
COPY build/ /usr/local/bin/

# copy project configuration files
COPY .andai.*.yaml /app/

# Set working directory
WORKDIR /app

SHELL ["/bin/zsh", "-c"]
ENTRYPOINT ["andai"]
