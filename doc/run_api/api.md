# How to install

## Make

Make is used for running commands related to building, running and debugging the API.

```bash
sudo apt install build-essentials
```

## Docker

The API is designed to run in a Docker container. The image is built whenever a new version is pushed to GitHub but inorder to run the built image, you will need docker installed on your computer/server.

To install docker engine and other related tools, see [Docker's official instructions](https://docs.docker.com/engine/install/debian/).

## Database

For instructions about installing and setting up new database, see `doc/local_db/postgresql.md`.

# How to run

## Starting API

To pull and start running the already built images from Docker Hub, type:

```bash
make compose-up
```

This will make sure that all necessary services are up and running.

## Stopping API

To stop running and clean up the stopped containers, type:

```bash
make compose-down
```

## Testing API

To verify that the API is running as it should, type:

```bash
curl localhost:8080/products
```

You should see a long list of products appear as result. If no products are found there is either no product data in DB or there is some other problem with the API. To help debugging the API server, you can view the logs from the container by typing:

```bash
make log-api
```
