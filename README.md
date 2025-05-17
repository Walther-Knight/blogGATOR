# blogGATOR
boot.dev course content: Build a Blog Aggregator in Go

Required software for this application:
Go 1.24.2+
Postgresql 16.8+
POSIX compatible Linux
Goose v3.24.3 (go install github.com/pressly/goose/v3/cmd/goose@latest)

Installation steps:
1) Pull the software from GitHub
2) Compile local install in ~/go/bin by running "go install" from the location you pulled to
3) Perform initial setup as below
4) Run the program with blogGATOR (may require a terminal restart)

Initial setup:
1) Create a user in linux and set a password with passwd
2) Create a database in postgresql named gator (CREATE DATABASE gator;)
3) Create a user with the same username and password in postgresql (ALTER USER <username> PASSWORD '<password>';)
5) Create .gatorconfig.json in your home directory. The blogGATOR looks for that config file in the root of home.
6) Update .gatorconfig.json with the below:
{
    "db_url":"postgres://<username>:<password>@localhost:5432/gator?sslmode=disable",
    "current_user_name":""
    }
Replace <username> with the user created earlier and replace <password> with the password string. CAUTION: This is passed as a URL so escape any special characters.
7) goose postgresql "postgres://<username>:<password>@localhost:5432/gator?sslmode=disable" up
This will set up the necessary tables

Commands:
login <username> sets the current user
register <username> registers a new user
users <no arguments> shows a list of users and the current user
agg <time interval> runs aggregation on the specified interval. This will read subscribed feeds and update their contents in the local database.
addfeed <name URL> adds a feed with a display name
feeds <no argument> lists all feeds and associated usernames
follow <URL> follows a feed with the current user
following <no argument> lists all feeds and followers
unfollow <URL> unfollows a feed for current user
browse <number> lists the latest RSS items from followed feeds. Defaults to 2 feeds.
