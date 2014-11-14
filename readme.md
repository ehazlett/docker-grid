# Docker Grid
This is an experiment in a public Docker service.  Node containers are launched on multiple Docker hosts and then the standard Docker client can run containers that are placed on the grid.  Everything runs completely in Docker.  There are no external services needed.

# Grid Controller
This is a lightweight "discovery" and light orchestration service.  Basically it just knows what machines exist and how to schedule simple launch tasks that are ran (containers).

The controller also has a lightweight aggregation service where it will take all of the nodes and aggregate them into a single Docker "grid".  You can access the "grid" using the standard Docker client.

# Grid Node
This queries the client Docker daemon to execute containers.  It also reports basic metadata like client resource limits and generalized location.

# Running Containers
To run containers on the grid, you simply use the Docker client.  The difference is you target (-H) a grid controller.

All containers run on the grid have an environment variable injected to allow for simple "filtering" when the node reports.  It will only report containers running that have this variable.  That way your other containers are not reported.

# Security
There is very little security.  This is meant to be a public service.  However, with the container "filtering", the grid will only report containers that are run using the grid service.

# Usage
This is just an experiment so do not use in any production-like environment.

## Controller
Start a single controller.

`docker run -d -p 8080:8080 ehazlett/docker-grid controller`

## Node
Start one or more nodes.

Note: this will attempt to detect your machine IP (to properly show exposed ports) -- you can alternatively use `-i <IP>` to override -- then you do not need `--net=host`.

`docker run -d -v /var/run/docker.sock:/var/run/docker.sock --net=host ehazlett/docker-grid node -c http://<controller-host-or-ip>:8080`
