![Slack Trail](trail.jpg)

Keeps track of slackers progress on the "oregon" trail. Posts a message to a slack channel when a
new slack user is created, and when an existing slack user is deleted.

## Concept

This application is initialized by fetching a list of all users from slack, which is stored in
postgres. The next time it runs, it fetches the list again, and compares the slack list to the
postgres list.

Any users in the slack list, but not in the postgres list are new users. Any users
marked as deleted in the slack list, but not deleted in the postgres list, are deleted users. A
message is posted to slack for each new/deleted user.

## Packaging

### List deployed functions

```sh
$ serverless deploy list functions
```


### Deploy all

```sh
$ make deploy
```

### Deploy single function

```sh
$ make
$ serverless deploy function --function users
```

## TODO

- [ ] output option, like console instead of slack
- [ ] diff avatars
- [ ] diff is_admin
- [ ] pagination, or at least warn on getting close to 1000 users (page limit I think)

---

- [x] track status
- [x] generic update user function, right now we have to add a new function to update like status
  for example when we start caring about that, we should just mark if there are any changes to a
  user and update everything
- [x] scripted aws deploys
- [x] sentry integration
- [x] document sls commands
- [x] document emoji

## Development

Expects env vars:

- SLACK_TOKEN
- SENTRY_DSN
- DATABASE_URL (used for dev)
- SLACK_CHANNEL_ID (used for dev)
- PROD_DATABASE_URL
- PROD_SLACK_CHANNEL_ID

Optionally, set aws keys for serverless access

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY

### Run locally

```sh
$ serverless invoke local --function users
```

### Using direnv

```sh
# .envrc

export DATABASE_URL='postgres://localhost:5432/slack_trail?sslmode=disable'
export SLACK_CHANNEL_ID='<Some Channel ID>'

export PROD_DATABASE_URL='<prod url>'
export PROD_SLACK_CHANNEL_ID='<Some Channel ID>'

export SENTRY_DSN='<sentry dsn>'

export AWS_ACCESS_KEY_ID='key'
export AWS_SECRET_ACCESS_KEY='secret'

# Makes psql connect to slack_trail database by default
export PGDATABASE='slack_trail'
```

### Database creation and setup

```sh
brew install golang-migrate
createdb slack_trail
migrate -path migrations -database 'postgres://localhost:5432/slack_trail?sslmode=disable' up
cd users && go run . init
cd emoji && go run . init
```

### New migration

```sh
migrate create -seq -dir migrations -ext sql some_file_name
```

### Migrate

```sh
# dev
migrate -path migrations -database 'postgres://localhost:5432/slack_trail?sslmode=disable' up
pg_dump -s slack_trail > structure.sql

# prod
migrate -path migrations -database "$PROD_DATABASE_URL" up
```

### Helpers

```sh
curl -sH "Authorization: Bearer $SLACK_TOKEN" https://slack.com/api/users.list | jq .
curl -sH "Authorization: Bearer $SLACK_TOKEN" https://slack.com/api/emoji.list | jq .
curl -sH "Authorization: Bearer $SLACK_TOKEN" https://slack.com/api/users.setPhoto -F image=@"/Users/zachtaylor/Downloads/slack-avatar.jpg"
curl -sH "Authorization: Bearer $SLACK_TOKEN" https://slack.com/api/users.deletePhoto
curl -sH "Authorization: Bearer $SLACK_TOKEN" https://slack.com/api/groups.info -F channel=GJUF0HLUC | jq -r '.group.members[]'
```
