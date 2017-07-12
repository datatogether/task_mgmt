# Task Management

Getting started:

If you have docker, clone this repo & run `docker-compose up`.

You should then be able to run the following:
```shell
  # this should respond with json, having an empty "data" array
  http://localhost:8080/tasks

  # this should respond with json, with meta.message : "task successfully enqueud"
  http://localhost:8080/ipfs/add?url=https://i.redd.it/5kwih5n5i58z.jpg

  # requesting this again should now show a task in the data array, including a "succeeded" timestamp:
  http://localhost:8080/tasks

  # congrats, you've put this url of a jpeg on ipfs: https://i.redd.it/5kwih5n5i58z.jpg
  # view it here:
  https://ipfs.io/ipfs/QmeDchVWNVxFcEvnbtBbio88UwrHSCqEAXpcj2gX3aufvv

  # connect to your ipfs server here:
  # click the "files" tab, and you'll see this hash: QmeDchVWNVxFcEvnbtBbio88UwrHSCqEAXpcj2gX3aufvv
  # this means you have a local ipfs node serving the image we just processed
  https://localhost:5001/webui
```

## From a recent Pull Request:

_TODO - make a proper readme, this is ripped from a recent Pull Request:_

tasks are any kind of work that needs to get done, but specifically work that would take longer, than say, a web request/response cycle should take. An example of a task can be "put this url on IPFS". another might be "identify the filetype of these 30,000 files, putting the results in a database".

Because this work will take anywhere from milliseconds to days, and may require special things to do that work, it makes sense to put those tasks in a queue, which is just a giant, rolling list of things to do, and have different services be able to add tasks to the queue, and sign up to do tasks as they hit the queue. This PR introduces a first-in-first-out (FIFO) queue to start with, meaning the first thing to get added is the first thing to get pulled off a list.

The queue itself is a server, specifically a rabbitmq sever, it's open source, and based on the open amqp protocol. This means that things that work with the queue don't necessarily need to be written in go. More on that in the future.

The task-mgmt service does just what it says on the tin. It's main job is to manage not just tasks, but the state of tasks as they move through the queue, questions like "what tasks are currently running?" are handled with this PR. As tasks are completed task-mgmt updates records of when tasks started, stopped, etc.

this PR removes all user interfaces and instead introduces both a JSON api and an remote procedure call (RPC) api, the RPC api will be used to fold all of task-mgmt into the greater datatogether api. I know, that's the word api a million times, basically this means we'll have a PR on the datatogether/api to expose tasks so that outside users will access tasks the same way they access, say, coverage, or users. Only internal services will need to use the task-mgmt JSON api as a way of talking across languages.

All of these changes turn the task-mgmt server into a backend service, so that we can fold all task stuff into the existing frontend. This means once the UI is written you'll be able to view, create, and track the progress of tasks from the standard webapp. PR on datatogether/context to follow.

Along with tracking tasks, task-mgmt both add to and reads from the queue. This might seem strange, but it makes for a more simple starting point. Later on down the road lots of different services may be registered to accept tasks from the queue, at which point we can transition task-mgmt to a role of just adding to the queue and tracking progress.

But most importantly of all, this PR also introduces a new subpackage task-mgmt/tasks which outlines the initial interface for a task definition, which is the platform on which tasks can be extended to all sorts of things. Getting this interface right is going to take some time, so I'd like to take some time to write an initial round of task-types, and then re-evaluate the interface.