#!/bin/bash

ENV_FILE="./data/secrets.env"

# Create or overwrite the .env file
touch $ENV_FILE

# Write environment variables to the .env file in one block
{
  echo "OPENAI_TOKEN=$OPENAI_TOKEN"
  echo "GEMINI_TOKEN=$GEMINI_TOKEN"
  echo "DATABASE_URL=$DATABASE_URL"
  echo "PORT=$PORT"
} >> $ENV_FILE
