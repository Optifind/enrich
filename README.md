# Enrich

## Installation

### Prerequisites

#### Go

> You need to have Go compiler installed and GOPATH setup correctly. Follow [these](https://go.dev/wiki/SettingGOPATH) instructions on setting up GOPATH.
> 
> It is recommended to use version manager for installing Go. For Mac and UNIX systems, use [this](https://github.com/moovweb/gvm) version manager.


---

Clone the repo to local directory and install dependencies

```bash
git clone git@github.com:lattots/enrich.git
cd enrich
go mod tidy
```

### Database initialization

#### PostgreSQL

> You need to have a running postgres server on your machine. You also have to have created a database and a user with all privileges to that database. For database installation and initialization see instructions [here](doc/local_db/postgresql.md).

---

### Updating secrets

> In order to run any scripts, you need to have a `secrets.env` file in `data/` folder. To create this file, create a copy of `secrets_template.env` in the same directory and rename it to `secrets.env`.
> 
> Once you have created the file, update all values in it to your own values. 

---

## Indexing products

Indexing products means creating product database from a CSV file and adding style and use-case embeddings for all products. This needs to be done before running product API.

Before running indexing algorithm, make sure to update `data/config.json` file so that `catalog-filepath` points to the product CSV file.

To run indexing algorithm, run:

```bash
make init
```

This creates a PostgreSQL database table in your local postgres server. The table will contain all products and their embeddings. The script also creates a database dump file that can later be used to set up product database in other environments such as Heroku.

---

## Running product API

### Prerequisites

Before you can run the API, you need to have set environment variable `ENVIRONMENT=local`. This allows the program to automatically differentiate between local and "production" environments and handle other environment variables accordingly.

On UNIX and Mac run:

```bash
export ENVIRONMENT=local
```

On Windows run:

```shell
set ENVIRONMENT=local
```

---

To run product API you first need to have run the indexing algorithm. Then you can simply start the API by running:

```bash
make api
```

### Test API

You can test local API by running in another terminal:

```bash
curl http://localhost:8080/products
```

The output should be a JSON with product information inside.

---

## Deploying to Heroku

To deploy a new version of product API or new product database, simply merge changes to branch `main`. This will automatically re-deploy the app in Heroku with the new product database.
