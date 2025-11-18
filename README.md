# Aggre-Gator
A blog aggregator written in Go. Last boot.dev guided project.

## Requirements
- Go 1.25+ or later
- PostgreSQL v15 or later

## Installation
Install through git-clone :
```bash
git clone github.com/Diwice/gator.git

cd gator

go build .
```
Then you will have to create a database for the app :
```bash
psql -c "CREATE DATABASE gator;"
```
Last step is to create a config in your home directory and edit it (~/.gatorconfig.json), it should contain a working db string in the following format :
```bash
{
    "db_url": "postgres://<user>:<password>@localhost:5432/gator"
}
```
you have to insert own user data (default user is postgres, and there is no password).

## Usage
After building the app, use it with any of the following commands :
- "login <user_name>" to change the current user
- "register <user_name>" to register a new user 
- "reset" to remove all entries from db
- "users" to display all users and the current user
- "agg <time>" to aggregate the posts with given time range between requests
- "addfeed "<feed_name>" "<feed_url>"" to add a new feed entry
- "feeds" to display all feeds
- "follow <feed_url>" to follow the feed from current user
- "following" to display followed feeds as current user
- "unfollow <feed_url>" to unfollow the feed as current user
- "browse <limit>" to browse the aggregated posts from followed feeds. Limited to 2 if not provided
Example :
```
./gator register Cathy
```
