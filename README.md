# gsevent

## Description
GSEvent is analytics aggregator. It stores different events coming from mobile SDKs into Redis and writes them 
into files for further processing.

Detailed documentation is available in _swagger.yaml_ file in the root of the repo.

## Requirements
1) [Docker](https://docs.docker.com/)
2) [Go](https://golang.org/doc/install)

## Setup

1) Install the necessary dependencies and libraries

    ```
    $ make init
    ```

2) Generate necessary the code

    ```
    $ make generate-swagger
    ```

3) Start the application
    
    ```
    $ docker-compose up -d
    ```

To init and start the service with one command:
    
    $ make run

