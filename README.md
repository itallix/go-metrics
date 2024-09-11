# go-metrics

My pet project to learn Golang with Yandex Practicum [Advanced Go Developer](https://practicum.yandex.ru/go-advanced/) course.

## Overview

`go-metrics` is a robust, scalable metrics collection system designed for distributed environments. It consists of two main components: an agent that collects and sends metrics in batches, and a server that receives, processes, and exposes these metrics through a RESTful API.

## Key Features

- Efficient Metric Collection: Agents collect system and application metrics with minimal overhead.
- Batch Processing: Metrics are sent in batches to optimize network usage and server processing.
- Scalable Architecture: Designed to handle multiple agents sending data to a centralized server.
- RESTful API: Server exposes collected metrics through a REST API.

## Components
### Agent

- Lightweight process that runs on target systems
- Collects various system metrics (CPU, memory, disk usage, network stats, etc.)
- Aggregates metrics and sends them in configurable batches (with assymetric encryption support)
- Implements retry logic

### Server

- Centralized metrics receiver and processor
- Efficiently handles incoming batch metrics from multiple agents
- Can store metrics in memory (with filesystem synchronization) or in PostgreSQL
- Provides a RESTful API for querying and analyzing metrics

## REST API Endpoints

_TBD_

## Tech Stack

_TBD_

## Development

### To get updates from template

Add git remote:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Update autotests:

```
git fetch template && git checkout template/main .github
```

### Requirements to run autotests

- Branch name should follow pattern `iter<number>`, where `<number>` — number of increment. E.g., branch `iter4` will trigger autotests for increments 1-4.
