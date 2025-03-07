# Quick Start

We will create new `andai` project config folder in `/tmp/test`

```bash
mkdir /tmp/test
touch /tmp/test/.andai.aider.yaml
touch /tmp/test/.andai.project.yaml
touch /tmp/test/docker-compose.yaml
touch /tmp/test/.redmine.env

➜  test tree -a /tmp/test
/tmp/test
├── .andai.aider.yaml
├── .andai.project.yaml
├── docker-compose.yaml
└── .redmine.env

0 directories, 4 files
```

Follow it up with copy & paste file contents from `/docs/examples/basic` folder.

That config is using anthropic API key with sonnet model.

## Target project

Let's create our first project that `andai` will work with. *You can use real project of course* and skip this step.
```
mkdir /tmp/test-repo
cd /tmp/test-repo
echo "# Test repo" > README.md
git init
git add README.md
git commit -m "Initial commit"
```

## Start Ticketing system

Now that you have configuration files in place, you can start ticketing system.

```bash
cd /tmp/test
docker-compose up -d
```

This will create new redmine (ticketing system) instance with database.

*(!) Do not configure it. `AndAI` will handle it in next step.*

It will take few seconds until redmine is up and running.

## andai binary
Copy `andai` binary there as well or add it to PATH so its available from everywhere.

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
cd /tmp/test/
andai lets go
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
rm -rf /tmp/test
rm -rf /tmp/test-repo
```
Be aware that `docker-compose down` will destroy images. If you want to stop and continue then use `docker-compose stop`.

## Follow up

Now that you created simple setup, and it is working as expected, 
you probably want to implement more complex workflows and add more real projects into the mix.

See other examples in [/docs/example](../examples/) folder and [workflow/README.md](workflow/README.md) documentation to see what is available.
