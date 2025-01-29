PROJECT=test make stop
PROJECT=test make start
PROJECT=test go run main.go issue create "Task" "Init repository" "Create README.md and main.py files. No content, just files."
PROJECT=test make work
