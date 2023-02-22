# Introduction

This tour introduces you to Compass. Along the way you will learn how to ingest assets in Compass database, list, search and manage asset lineage.
### Prerequisites

This tour requires you to have a Compass CLI tool installed on your local machine. 
You can run `compass` to verify the installation. Please follow [installation](../installation) and [configuration](../configuration) guides if you do not have it installed already.

Compass CLI and clients talks to Compass server to publish and fetch assets, search and lineage. Please make sure you also have a Compass server running. You can also run server locally with `compass server start` command. For more details check deployment guide.

### Help

At any time you can run the following commands.

```
# See the help for a command
$ compass --help
```

The list of all available commands are as follows:

```text
Core commands
  asset          Manage assets
  discussion     Manage discussions
  lineage        observe the lineage of metadata
  search         query the metadata available

Other commands
  completion     Generate shell completion scripts
  config         Manage server and client configurations
  help           Help about any command
  server         Manage server
  version        Print version information

Help topics
  environment    List of supported environment variables
  reference      Comprehensive reference of all commands
```

Help command can also be run on any sub command with syntax `compass <command> <subcommand> --help` Here is an example for the same.

```
$ compass asset --help
```

### Background for this tutorial

Let's imagine we have a postgres instance running with a database called `my-database` that has plenty of tables. One of the tables is named `orders`. We will ingest this asset, list and search the metadata from Compass. If you don't know what an asset is, don't worry we have got you covered in the [next page](./1-my-first-asset.md#12-hello-world-asset). We will also be defining certain rules to inserting and quering lineage between tables `dailyorders` and `orders` in this example guide.  