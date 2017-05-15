package main

const qAvailableTasks = `
WITH t AS (
  SELECT
    repos.url as repo_url,
    repos.latest_commit as repo_commit,
    sources.url as source_url,
    sources.checksum as source_checksum
  FROM sources, repos, repo_sources
  WHERE 
    sources.id = repo_sources.source_id AND
    repos.id = repo_sources.repo_id
)
SELECT
  t.repo_url, t.repo_commit, t.source_url, t.source_checksum
FROM t LEFT OUTER JOIN tasks ON (t.source_url = tasks.source_url)
WHERE
  tasks.repo_commit is null OR
  tasks.source_checksum is null;`

const qTasksBySourceUrl = `
SELECT
  id, created, updated, request, success, fail, 
  code_url, code_commit, source_url, source_checksum, result_url, result_hash, message
FROM tasks
ORDER BY $1
LIMIT $2 OFFSET $3;`

const qTaskReadById = `
SELECT 
  id, created, updated, request, success, fail, 
  code_url, code_commit, source_url, source_checksum, result_url, result_hash, message
FROM tasks
WHERE id = $1;`

const qTaskInsert = `
INSERT INTO tasks
  (id, created, updated, request, success, fail, 
  code_url, code_commit, source_url, source_checksum, result_url, result_hash, message)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`

const qTaskUpdate = `
UPDATE tasks SET
  created = $2, updated = $3, request = $4, success = $5, fail = $6, 
  code_url = $7, code_commit = $8, source_url = $9, source_checksum = $10, result_url = $11, result_hash = $12, message = $13
WHERE id = $1;`

const qTaskDelete = `DELETE FROM tasks WHERE id = $1;`

const qSourcesBySourceUrl = `
SELECT
  id, created, updated, name, url, checksum, meta
FROM sources
ORDER BY $1
LIMIT $2 OFFSET $3;`

const qSourceReadById = `
SELECT 
  id, created, updated, name, url, checksum, meta
FROM sources
WHERE 
  id = $1;`

const qSourceInsert = `
INSERT INTO sources
  (id, created, updated, name, url, checksum, meta)
VALUES
  ($1, $2, $3, $4, $5, $6, $7);`

const qSourceUpdate = `
UPDATE sources SET
  created = $2, updated = $3, name = $4, url = $5, checksum = $6, meta = $7
WHERE id = $1;`

const qSourceDelete = `DELETE FROM sources WHERE id = $1;`

const qReposBySourceUrl = `
SELECT
  id, created, updated, url, branch, latest_commit
FROM repos
ORDER BY $1
LIMIT $2 OFFSET $3;`

const qRepoReadById = `
SELECT 
  id, created, updated, url, branch, latest_commit
FROM repos
WHERE id = $1;`

const qRepoInsert = `
INSERT INTO repos
  (id, created, updated, url, branch, latest_commit)
VALUES
  ($1, $2, $3, $4, $5, $6);`

const qRepoUpdate = `
UPDATE repos SET
  created = $2, updated = $3, owner = $4, name = $5, branch = $6, latest_commit = $7
WHERE id = $1;`

const qRepoDelete = `DELETE FROM repos WHERE id = $1;`
