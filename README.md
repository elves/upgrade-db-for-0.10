# Script to convert Elvish db from SQLite format to BoltDB format.

## Install

go get github.com/elves/upgrade-db-for-0.10

## Usage

1. stop all elvish session, including daemon.
2. rename current db:
   mv ~/.elvish/db ~/.elvish/db.sqlite
3. convert db to new format:
   $GOPATH/bin/upgrade-db-for-0.10 ~/.elvish/db.sqlite ~/.elvish/db
