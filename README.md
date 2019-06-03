docker-backup
=============

A tool to create & restore complete, self-contained backups of Docker containers

# What's the issue

Docker services usually have a bunch of volatile data volumes that need to be
backed up. Backing up an entire (file)system is easy, but often enough you just
want to create a backup of a single (or a few) containers, maybe to restore them
on another system later.

Some services also need to be aware (flushed/synced/paused) of an impending
backup, e.g. databases. The backup should be run on the Docker host, as you
don't want to have a backup client configured & running in every single
container either, since this would add a lot of maintenance & administration
overhead.

`docker-backup` directly connects to Docker, analyzes a container's mounts &
volumes, and generates a list of dirs & files that need to be backed up on the
host system. This also collects all the metadata information associated with a
container, so it can be restored or cloned on a different host, including its
port-mappings and data volumes.

The generated list can either be fed to an existing backup solution or
`docker-backup` can directly create a `.tar` image of your container, so you can
simply copy it to another machine.

## Installation

`docker-backup` requires Go 1.11 or higher. Make sure you have a working Go
environment. See the [install instructions](http://golang.org/doc/install.html).

`docker-backup` works with Docker hosts running Docker 18.02 (API version 1.36)
and newer.

### Packages

- Arch Linux: [docker-backup](https://aur.archlinux.org/packages/docker-backup/)

### From source

    git clone https://github.com/muesli/docker-backup.git
    go build

Run `docker-backup --help` to see a full list of options.

## Usage

### Creating a Backup

To backup a single container start `docker-backup` with the `backup` command and
supply the ID of the container:

    docker-backup backup <container ID>

This will create a `.json` file with the container's metadata, as well as a file
containing all the volumes that need to be backed up with an external tool like
restic or borgbackup.

If you want to directly create a `.tar` file containing all the container's
data, simply run:

    docker-backup backup --tar <container ID>

You can also backup all running containers on the host with the `--all` flag:

    docker-backup backup --all

To backup all containers (regardless of their current running state), run:

    docker-backup backup --all --stopped

With the help of `--launch` you can directly launch a backup program with the
generated file-list supplied as an argument:

    docker-backup backup --all --launch "restic -r /dest backup --password-file pwfile --tag %tag --files-from %list"

### Restoring a Backup

To restore a container, run `docker-backup` with the `restore` command:

    docker-backup restore <backup_file>

`docker-backup` will automatically detect whether you supplied a `.tar` or
`.json` file and restore the container, including all its port-mappings and data
volumes.

If you want to start the container once the restore has finished, add the
`--start` flag:

    docker-backup restore --start <backup_file>

## Development

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/muesli/docker-backup)
[![Build Status](https://travis-ci.org/muesli/docker-backup.svg?branch=master)](https://travis-ci.org/muesli/docker-backup)
[![Go ReportCard](http://goreportcard.com/badge/muesli/docker-backup)](http://goreportcard.com/report/muesli/docker-backup)
