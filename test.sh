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

PROJECT=test build/andai issue move "Init repository" True
#PROJECT=test bin/andai move-children # will move created children from Initial to Backlog
