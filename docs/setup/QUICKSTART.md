# Quick Start

Please choose a project folder to be used as config and data storage purposes. It will be the main directory with all execution happens.

Please make sure that this path is not under exisitng git repo beause we need to init a new git repo here.

We will create new `andai` project config folder in `/tmp/test` (use any folder you like),
copy configuration files from [config_template/](../config_template/) folder and edit them.

`test` is the root folder for configuration and `test-repo` will be used as a git repository for code generation.

```bash
mkdir /tmp/test ; cd /tmp/test # or any other folder you like

mkdir /test-repo ; cd /test-repo # or any other folder you like

git init

echo "This is the code repo" > README.md

git commit -am "Created with quickstart script"

# Download configuration files
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/.andai.aider.yaml
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/.andai.project.yaml
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/docker-compose.yaml
wget https://raw.githubusercontent.com/andrejsstepanovs/andai/refs/heads/main/docs/config_template/.redmine.env

# Check that files are there
➜  ls -la /tmp/test
/tmp/test
├── .andai.aider.yaml
├── .andai.project.yaml
├── docker-compose.yaml
└── .redmine.env

0 directories, 4 files
```

Then (!) Edit these 3 files:
- `.andai.aider.yaml`
- `.andai.project.yaml`
- `docker-compose.yaml`

## Start Ticketing system

Now that you have configuration files in place (and adjusted with your project and llm config), you can start ticketing system.

```bash
cd /tmp/test
docker-compose up -d
```

This will create new redmine (ticketing system) instance with database.

*(!) Do not configure it. `AndAI` will handle it in next step.*

It will take few seconds until redmine is up and running.

## andai binary
Copy `andai` binary there as well or add it to PATH so it's available from everywhere.

## Configure ticketing system
Now it's time to start using `AndAI` binary.
There are multiple commands that are focusing on setup tasks and ping commands to make sure that all is configured correctly and ready for work.

```bash
cd /tmp/test/

# check that .andai.project.yaml is valid
andai validate config

# this command will only set things up
andai setup all

# this command will ping services. useful to check if everything is in order.
andai ping all

# run single work loop
andai work next

# never ending work loop
andai work loop
```

Alternative (all in one):
```bash
# from same folder where you have `.andai.project.yaml`
andai go
```

## Create Ticket

Open Redmine in browser. If you used provided configuration files, it should be available at `http://localhost:10083`.

With username: `admin`, Password: `admin`.

We do not care about ACL and other security issues, because this is local setup that you in full control.

Let's create simple ticket like:
```
Improve README.md documentation
```

After that observe terminal command that is running `andai`. It should pick up this ticket and start working on it.

## Stop and cleanup
```bash
docker-compose down
```
Be aware that `docker-compose down` will destroy images. If you want to stop and continue then use `docker-compose stop`.

## Follow up

Now that you created simple setup, and it is working as expected, 
you probably want to implement more complex workflows and add more real projects into the mix.

See other examples in [/docs/workflow_examples](../workflow_examples/) folder and [workflow/README.md](workflow/README.md) documentation to see what is available.
