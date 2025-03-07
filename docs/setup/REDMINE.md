# redmine

Redmine is a flexible project management web application. Written using the Ruby on Rails framework, it is cross-platform and cross-database.

This part of confuguration is quite simple. Most of the values are hardcoded and mean nothing.

- db - Database connection. Because redmine is running in docker-compose, this value should be aligned with `.redmine.env` and `docker-compose.yaml` `database` setup.
- url - Redmine URL. Same story as with `db`.
- api_key - Redmine API key. Hardcode to anything you want really. We are sticking with single `admin` user. If you're running this locally, you can stick with this value.
- repositories - Path to repositories from where redmine is (container). Make sure that in `docker-compose.yaml` your project repositories are mounted to this path.

```yaml
redmine:
  db: redmine:redmine@tcp(localhost:3306)/redmine
  url: "http://localhost:10083"
  api_key: "2159cef2fb6c82c4f66981f199798781e161c694"
  repositories: "/var/repositories/"
```

`repositories` are not mandatory to be correct. But if correct you will have access to your project repositories from redmine web ui.
