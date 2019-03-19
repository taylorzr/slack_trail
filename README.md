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
Use bin/build_lambda.sh to package for lambda, then upload zip. TODO: Use aws cli to update lambda
automatically

## Todo
- [ ] pagination, or at least warn on getting close to 1000 users (page limit I think)
- [x] scripted aws deploys
- [x] sentry integration

## Development
Expects env vars:
- DATABASE_URL (currently using elephantsql db)
- SLACK_TOKEN
- SLACK_CHANNEL_ID

### Database creation and setup

```
create_db avant_trail
migrate -path migrations -database $DATABASE_URL up
go run . init
```

### New migration
```
migrate create -seq -dir migrations -ext sql some_file_name
```

### Migrate
```
migrate -path migrations -database $DATABASE_URL up
pg_dump -s avant_trail > structure.sql
```

### Helpers
curl -H "Authorization: Bearer $SLACK_TOKEN" https://slack.com/api/users.list | jq
