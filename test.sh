PROJECT=test make build
PROJECT=test make rm

rm -rf /tmp/test
mkdir /tmp/test
cd /tmp/test
git init
touch .gitignore
git add .gitignore
git commit -m 'init'
cd -

PROJECT=test make start
PROJECT=test build/andai issue create "Task" "Init repository" "Create README.md and main.py files. No content, just files."
PROJECT=test build/andai issue move "Init repository" True

PROJECT=test make work # will move to Analysis
PROJECT=test make work # will work on Analysis and move to In Progress
PROJECT=test make work # will create 2 Step tasks and put them in Initial, finally wil move Task to Testing
PROJECT=test make work # will move Task from QA

PROJECT=test build/andai issue move "Init repository" "Approved" # move parent task to Approved
PROJECT=test build/andai work triggers # Will move children to Backlog
PROJECT=test make work # will move children to Analysis (if not dependent)

PROJECT=test make work # will work on 1 child from Analysis and move to In Progress
PROJECT=test make work # will move task to QA
PROJECT=test make work # will move task from QA to Approved


PROJECT=test make work # Will move to Approved
PROJECT=test make work # Will move to Deployment
PROJECT=test make work # Will merge to parent and move to Done

# now other task will be unblocked
PROJECT=test make work # Will move to Analysis
PROJECT=test make work # will work on Analysis and move to In Progress
PROJECT=test make work # Will work on In Progress and move to Testing
PROJECT=test make work # Will move from Testing to QA
PROJECT=test make work # Will move from QA to Approved
PROJECT=test make work # Will move to Deployment
PROJECT=test make work # Will merge to parent and move to Done. Then should figure out that all children of Task are done and move Task from Approved to Deployment

PROJECT=test make rm
