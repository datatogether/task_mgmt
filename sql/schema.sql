-- name: drop-all
DROP TABLE IF EXISTS tasks, sources, repos, repo_sources;

-- name: create-tasks
CREATE TABLE tasks (
  id               UUID NOT NULL PRIMARY KEY,
  created          timestamp NOT NULL DEFAULT (now() at time zone 'utc'),
  updated          timestamp NOT NULL DEFAULT (now() at time zone 'utc'),
  title            text NOT NULL DEFAULT '',
  request          timestamp,
  success          timestamp,
  fail             timestamp,
  repo_url         text NOT NULL DEFAULT '',
  repo_commit      text NOT NULL DEFAULT '',
  source_url       text NOT NULL DEFAULT '',
  source_checksum  text NOT NULL DEFAULT '',
  result_url       text NOT NULL DEFAULT '',
  result_hash      text NOT NULL DEFAULT '',
  message          text NOT NULL DEFAULT ''
);

-- name: create-sources
CREATE TABLE sources (
  id               UUID NOT NULL PRIMARY KEY,
  created          timestamp NOT NULL DEFAULT (now() at time zone 'utc'), 
  updated          timestamp NOT NULL DEFAULT (now() at time zone 'utc'), 
  title            text NOT NULL DEFAULT '',
  url              text NOT NULL,
  checksum         text NOT NULL DEFAULT '', 
  meta             json
);

-- name: create-repos
CREATE TABLE repos (
  id               UUID NOT NULL PRIMARY KEY,
  created          timestamp NOT NULL DEFAULT (now() at time zone 'utc'), 
  updated          timestamp NOT NULL DEFAULT (now() at time zone 'utc'), 
  url              text NOT NULL, 
  branch           text NOT NULL default 'master',
  latest_commit    text NOT NULL
);

-- name: create-repo_sources
CREATE TABLE repo_sources (
  repo_id          UUID NOT NULL references repos(id) ON DELETE CASCADE,
  source_id        UUID NOT NULL references sources(id) ON DELETE CASCADE
);