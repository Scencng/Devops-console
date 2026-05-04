# DevOps Console

This repository contains the complete DevOps platform project, including the main backend and frontend, the Kafka console subproject, and the integrated MySQL management project sources.

## Project layout

- `backend/`: Go backend services for the main DevOps platform
- `frontend/`: Vue frontend application for the main DevOps platform
- `kafka-console/`: standalone Kafka console project
- `mysql-console/`: MySQL visual management project source
- `README.md`: repository overview
- `.gitignore`: repository ignore rules

## Notes

- The MySQL module has been integrated into the main DevOps platform under `frontend/src/mysql/` and related backend MySQL routes and services
- Local runtime artifacts such as build output, browser automation output, and dependency directories are excluded from version control
