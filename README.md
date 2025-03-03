This is a simple CLI that generates traffic to a server based on a specified rate of requests/second, duration, and target url.

It currently has two modes: 
1. Base mode, which generates a consistent traffic load at a desired rate of requests/second and duration. 
2. Random mode, which generates traffic with random spikes of higher traffic.

There's also some kubernetes functionality that is currently under development.

This includes features such as:
Replica Distribution: Track how requests are distributed amongst replicas.
Auto-scaling: Receive notifications when new replicas are created or deleted based on traffic demand.
Resource Limit Monitoring: Get alerts when your service reaches its resource limits (e.g., CPU or memory)

The goal of this is to explore how different deployment, rollout and service configurations behave under various traffic loads.
