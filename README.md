# Scan Server

This project will not likely be of much use generally but the code is available here for you to study or borrow from if 
you find value. This project implements several components that take the number crunching from a private project and
divides the work so that it can scale to cloud workloads.

It looks something like this:
![Cloud Diagram](http://pasta.e8.particle-explorer.s3-website-us-west-2.amazonaws.com/Cloud%20Scanner%20-%20Scan%20Service.png)

The client is developed in Unity (private repo). It uses the GPU and compute shaders to do a lot of number crunching in 
parallel and allows you to view and explore the results. The goal of this project is to dramatically increase the speed and
surface area of the scan space by moving the compute operations into the cloud.

The Unity client sends a JSON message to the Gateway. The Gateway processes the message and potentially generates many 
work request messages, dropping them onto the message bus. Any number of large compute + GPU instances pick up a request and 
crunch on it for the amount of work specified in the message.

Each message includes a minimum bar that a result must meet to continue.  Results that qualify are dropped onto the message 
bus and further processed. A Persistence service picks up a copy of these results and serializes them to a database. Separately, 
one or more QoS services (quality of service) pick up the results, buffer them, sort them ordered by the "best" results 
and enqueues them and a "best effort" sorted queue back to the Unity client to visualize the progress.

## Installation
TODO

## Usage
TODO
