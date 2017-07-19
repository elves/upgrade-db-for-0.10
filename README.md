# Script to convert Elvish db from SQLite format to BoltDB format.

## Install

```
go get github.com/elves/upgrade-db-for-0.10
```

## Usage

After upgrading Elvish to the latest commit:

1. Kill the daemon: `kill $daemon:pid`
2. Rename current db:
   `mv ~/.elvish/db ~/.elvish/db.sqlite`
3. Convert db to new format: `upgrade-db-for-0.10 ~/.elvish/db.sqlite ~/.elvish/db`
4. Restart daemon: `daemon:spawn`
