# Docker Grid
This is an experiment in a public Docker service.  Node containers are launched on multiple Docker hosts and then the standard Docker client can run containers that are placed on the grid.

# Grid Controller
This is a lightweight "discovery" and light orchestration service.  Basically it just knows what machines exist and how to schedule simple launch tasks that are ran (containers).

The controller also has a lightweight aggregation service where it will take all of the nodes and aggregate them into a single Docker "grid".  You can access the "grid" using the standard Docker client.

# Grid Agent
This is a simple Go application that queries the client Docker daemon to execute containers.  It also reports basic metadata like client resource limits and generalized location.

# Running Containers
To run containers on the grid, you simply use the Docker client.  The difference is you target (-H) a grid controller.

All containers run on the grid have an environment variable injected to allow for simple "filtering" when the node reports.  It will only report containers running that have this variable.  That way your other containers are not reported.

# Security
There is very little security.  This is meant to be a public service.  However, with the container "filtering", the grid will only report containers that are run using the grid service.
