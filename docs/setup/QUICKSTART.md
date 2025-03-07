# Quick Start

Quick start guide to get you up and running.

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

## Find a place for AndAI configurations

Figure out location where you will be running andai from. It will require config files to be located there.
It is recommended to create new small local git repository for this purpose (not mandatory).

Create new folder with files (copy contents of these files from `/docs/examples`):
- `docker-compose.yml`
- `.redmine.env`
- `.andai.project.yaml`
- `.andai.aider.yaml`

## Start Ticketing system

Now that you have configuration files in place, you can start ticketing system.
`cd` into this folder and run `docker-compose up -d`.

This will create new redmine (ticketing system) instance with database.

*(!) Do not configure it. `AndAI` will handle it in next step.*

## Configure ticketing system

Now it's time to start using `AndAI` binary.
There are multiple commands that are focusing on setup tasks and ping commands to make sure that all is configured correctly and ready for work.

```bash
# this command will set everything up and start the work on tickets
andai lets go

# this command will only set things up
andai setup all

# this command will ping services. useful to check if everything is in order.
andai ping all

# run single work loop
andai work next

# never ending work loop
andai work loop
```

## Create Ticket

Open Redmine in browser. If you used provided configuration files, it should be available at `http://localhost:10083`.

With username: `admin`, Password: `admin`.

We do not care about ACL and other security issues, because this is local setup that you in full control.

Let's create simple ticket like:
```
Improve README.md documentation.
```

After that observe terminal command that is running `andai`. It should pick up this ticket and start working on it.

## Follow up

Now that you created simple setup, and it is working as expected, 
you probably want to implement more complex workflows and add more real projects into the mix.

See other examples in [/docs/example](../examples/) folder and [workflow/README.md](workflow/README.md) documentation to see what is available.
