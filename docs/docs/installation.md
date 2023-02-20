# Installation

There are several approaches to install Compass.

1. [Using a pre-compiled binary](#binary-cross-platform)
2. [Using the Docker image](#docker)
3. [Installing from source](#building-from-source)

### Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/compass/releases) page. Once downloaded, the binary can be run from anywhere.
You don’t need to install it into a global location. This works well for shared hosts and other systems where you don’t have a privileged account.
Ideally, you should install it somewhere in your PATH for easy use. `/usr/local/bin` is the most probable location.

#### macOS

`Compass` is available via a Homebrew Tap, and as downloadable binary from the [releases](https://github.com/odpf/compass/releases) page:

```sh
brew install odpf/tap/compass
```

To upgrade to the latest version:

```
brew upgrade compass
```

#### Linux

`Compass` is available as downloadable binaries from the [releases](https://github.com/odpf/compass/releases/latest) page. Download the `.deb` or `.rpm` from the releases page and install with `sudo dpkg -i` and `sudo rpm -i` respectively.

#### Windows

`compass` is available via [scoop](https://scoop.sh/), and as a downloadable binary from the [releases](https://github.com/odpf/compass/releases/latest) page:

```
scoop bucket add compass https://github.com/odpf/scoop-bucket.git
```

To upgrade to the latest version:

```
scoop update compass
```

### Docker

We provide ready to use Docker container images. To pull the latest image:

```
docker pull odpf/compass:latest
```

To pull a specific version:

```
docker pull odpf/compass:v0.3.2
```

If you like to have a shell alias that runs the latest version of compass from docker whenever you type `compass`:

```
mkdir -p $HOME/.config/odpf
alias compass="docker run -e HOME=/tmp -v $HOME/.config/odpf:/tmp/.config/odpf --user $(id -u):$(id -g) --rm -it -p 3306:3306/tcp odpf/compass:latest"
```

### Building from Source

Begin by cloning this repository then you have two ways in which you can build compass

* As a native executable
* As a docker image

Run either of the following commands to clone and compile Compass from source

```bash
$ git clone git@github.com:odpf/compass.git                 # (Using SSH Protocol)
$ git clone https://github.com/odpf/compass.git             # (Using HTTPS Protocol)
```
#### As a native executable

To build compass as a native executable, run `make` inside the cloned repository.

```bash
$ make
```

This will create the `compass` binary in the root directory

Initialise server and client config file. Customise the `compass.yml` file with your local configurations.

```bash
$ ./compass config init
```

Run database migrations

```bash
$ ./compass server migrate
```

Start compass server

```bash
$ ./compass server start
```

#### As a Docker image

Building compass' Docker image is just a simple, just run docker build command and optionally name the image

```bash
$ docker build . -t compass
```

### Verifying the installation​

To verify if Compass is properly installed, run `compass --help` on your system. You should see help output. If you are executing it from the command line, make sure it is on your PATH or you may get an error about Compass not being found.

```bash
$ compass --help
```

### What's next

- See the [Configurations](./configuration.md) page on how to setup Compass server and client
- See the [CLI Reference](./reference/api.md) for a complete list of commands and options.
