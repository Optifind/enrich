# Init Database

Init database is a script that initializes a new database according to the options in `data/config.json`. After running the script there will be a new product catalog collection in the Milvus database that has all relevant product information for later use.

Run the script by running `make init` in `backend/`.

Make sure to update `data/config.json` and `data/secrets.env` before running the script.