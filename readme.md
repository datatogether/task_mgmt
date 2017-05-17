# Task Management

This is a [pilot project in collaboration with IPFS](https://github.com/ipfs/distributed-wikipedia-mirror/issues/8) to assist initiating & tracking mirroring kiwix wikipedia dumps onto IPFS by way of excuting shell scripts.

Currently this project periodically checks a subset of [kiwix wikipedia downloads](http://wiki.kiwix.org/wiki/Content_in_all_languages) and [ipfs/distributed-wikipedia-mirror](https://github.com/ipfs/distributed-wikipedia-mirror) for changes. If the repo master changes or any specified project's md5 checksum changes, this server will generate a "task" that qualified users will be able to set into action. Users are qualified by connecting their github account to the archivers.space identity management server. Permission to initiate tasks are determined by weather the user has either `admin` or `write` access to the `ipfs/distributed-wikipedia-mirror` repository. When a qualified user hits `run` on a task, an email is sent to a specified set of recipients, who will receive links to follow when the migration either succeeds or fails (again, requiring qualified access). This email process is a temporary stopgap while we work out a fully automated pipeline for executing & tracking tasks.

### Roadmap:
* [x] Uh, Tests.
* [ ] Golint & general Cleanup
* [ ] Connect tasks to a que
* [ ] Automate task execution from the end of that que
* [ ] Repurpose email to only handle IPNS repoints, notifications
* [ ] Connect resulting IPFS hash & IPNS url to success handler
* [ ] Refactor & generalize, abstracting kiwix-specific state tracking to... somewhere else.
* [ ] Commit results of tasks back to the repository, updating `snapshot-hashes.yml`
* [ ] Allow users to manage which kiwix projects should be injested
* [ ] UI Cleanup
* [ ] Open up public display of task que, only presenting action buttons to qualified users.