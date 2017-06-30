package tasks

const qTaskCreateTable = `
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
);`

// an available task a source.Checksum && repo.LatestCommit combination that doesn't
// have a task model already created.
const qAvailableTasks = `
WITH t AS (
  SELECT
    repos.url as repo_url,
    repos.latest_commit as repo_commit,
    sources.title as source_title,
    sources.url as source_url,
    sources.checksum as source_checksum
  FROM sources, repos, repo_sources
  WHERE 
    sources.id = repo_sources.source_id AND
    repos.id = repo_sources.repo_id
)
SELECT
  t.repo_url, t.repo_commit, t.source_title, t.source_url, t.source_checksum
FROM t LEFT OUTER JOIN tasks ON (t.source_url = tasks.source_url)
WHERE
  tasks.repo_commit is null OR
  tasks.source_checksum is null;`

const qTasks = `
SELECT
  id, created, updated, title, request, success, fail, 
  repo_url, repo_commit, source_url, source_checksum, result_url, result_hash, message
FROM tasks
ORDER BY created DESC
LIMIT $1 OFFSET $2;`

const qTaskExists = `SELECT exists(SELECT 1 FROM tastsk WHERE id = $1);`

const qTaskReadById = `
SELECT 
  id, created, updated, title, request, success, fail, 
  repo_url, repo_commit, source_url, source_checksum, result_url, result_hash, message
FROM tasks
WHERE id = $1;`

const qTaskInsert = `
INSERT INTO tasks
  (id, created, updated, title, request, success, fail, 
  repo_url, repo_commit, source_url, source_checksum, result_url, result_hash, message)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`

const qTaskUpdate = `
UPDATE tasks SET
  created = $2, updated = $3, title = $4, request = $5, success = $6, fail = $7, 
  repo_url = $8, repo_commit = $9, source_url = $10, source_checksum = $11, result_url = $12, result_hash = $13, message = $14
WHERE id = $1;`

const qTaskDelete = `DELETE FROM tasks WHERE id = $1;`
