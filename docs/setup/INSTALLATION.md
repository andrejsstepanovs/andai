# Installation

## Download

Download `andai` executable from [github releases](https://github.com/andrejsstepanovs/andai/tags)

## Build

Checkout the repo and build it from source.

```bash
# building from source
git clone git@github.com:andrejsstepanovs/andai.git
cd andai
make build
ls -l ./build/andai
# add it to PATH or create alias to it
```

## Aider

If you run `andai` in docker, this step is not necessary.

If you plan to run `andai` locally, you will need to install [aider](https://aider.chat/).

Install it and make sure it is available in your PATH. No other configuration is necessary, 
as most of the `aider` configuration will be done via command line arguments.

```
uv tool install aider
uv tool upgrade aider-chat
```

## Quick Start

After this follow [Quick Start](QUICKSTART.md) guide to get you up and running.
