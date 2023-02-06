# Configuration

## Server Setup
Compass binary contains both the CLI client and the server. Each has it's own configuration in order to run. Server configuration contains information such as database credentials, elastic search brokers, log severity, etc. while CLI client configuration only has configuration about which server to connect. In order to run compass locally, you’ll need to have an instance of postgres and elasticsearch running.

#### Pre-requisites
- Postgres 13 or more
- Elastic Search v7.0

### Initialization
Create a compass.yaml file (`touch compass.yaml`) in the root folder of Compass project or [use `--config` flag](#using---config-flag) to customize to config file location, or you can also [use environment variables](#using-environment-variable) to provide the server configuration. 

Setup up a database in Postgres and provide the details in the DB field as given in the example below. For the purpose of this tutorial, we'll assume that the database name is `compass`, host and port for the database are `localhost` and `5432`. The server is running on `localhost` and port `8080`.

> If you're new to YAML and want to learn more, see [Learn YAML in Y minutes.](https://learnxinyminutes.com/docs/yaml/)

Following is a sample server configuration yaml:

```yaml
log_level: info                                 # debug|info|warning|error|fatal|trace|panic - default: info

elasticsearch:
    brokers: http://localhost:9200              #required

db:
    host: localhost                             #required
    port: 5432                                  #required
    name: compass                               #required
    user: compass                               #required
    password: compass_password                  #required
    sslmode: disable                            #optional

service:
    host: localhost                             #required
    port: 8080                                  #required    
    identity:                                   
        headerkey_uuid: Compass-User-UUID       #required
        headerkey_email: Compass-User-Email     #optional
        provider_default_name: shield           #optional
    grpc:
        port: 8081                              #required
        max_send_msg_size: 33554432     
        max_recv_msg_size: 33554432
```

#### Starting the server

Database migration is required during the first server initialization. In addition, re-running the migration command might be needed in a new release to apply the new schema changes (if any). It's safer to always re-run the migration script before deploying/starting a new release.

To initialize the database schema, Run Migrations with the following command:
```sh
$ compass migrate
```

To run the Compass server use command:

```sh
$ compass serve
```

### Using `--config` flag

```bash
$ compass migrate --config=<path-to-file> 
```

```bash
$ compass serve --config=<path-to-file>
```

### Using environment variable

All the server configurations can be passed as environment variables using underscore _ as the delimiter between nested keys. 
Here is the corresponding environment variable for the above

Configuration key       | Environment variable      |
------------------------|---------------------------|
ELASTICSEARCH.BROKERS   | ELASTICSEARCH_BROKERS     |
DB.HOST                 | DB_HOST                   |
DB.NAME                 | DB_NAME                   |
DB.PASSWORD             | DB_PASSWORD               |

Set the env variable using export
```bash
$ export DB_PORT = 5432
```


### Required Header/Metadata in API
Compass has a concept of [User](./concepts/user.md). In the current version, all HTTP & gRPC APIs in Compass requires an identity header/metadata in the request. The header key is configurable but the default name is `Compass-User-UUID`.

Compass APIs also expect an additional optional e-mail header. This is also configurable and the default name is `Compass-User-Email`. The purpose of having this optional e-mail header is described in the [User](./concepts/user.md) section.

## Client Initialisation

Add client configurations in the same `~/compass.yaml` file in root of current directory. Open this file to configure client. 

```yml
client:
    host: localhost:8081
    serverheaderkey_uuid: Compass-User-UUID
    serverheadervalue_uuid: john.doe@example.com
```