# PGS3

A collection of scripts to help manage PostgreSQL backups to S3. It offers a simple CLI to backup and restore PostgreSQL databases via `pg_dump` and `pg_restore` and upload the backups to an S3 bucket. It also offers a simple scheduler to automate the backups and cleanup old ones.

## Installation

This package was mainly developed for use as a Docker container within a docker-compose setup.

```yml
version: "3.8"

services:
  db:
    image: postgres:latest
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    ports:
      - "${DB_PORT}:5432"
    volumes:
      - db_volume:/var/lib/postgresql/data

  pgs3:
    image: "ghcr.io/zaniluca/pgs3:latest" # <--
    command: backup --schedule @midnight --keep-days 7
    environment:
      PGS3_AWS_REGION: ${S3_REGION}
      PGS3_AWS_ACCESS_KEY_ID: ${S3_ACCESS_KEY_ID}
      PGS3_AWS_SECRET_ACCESS_KEY: ${S3_SECRET_ACCESS_KEY}
      PGS3_AWS_S3_BUCKET: ${S3_BUCKET}
      PGS3_POSTGRES_HOST: db
      PGS3_POSTGRES_DB: ${POSTGRES_DB}
      PGS3_POSTGRES_USER: ${POSTGRES_USER}
      PGS3_POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    depends_on:
      db:
        condition: service_healthy
```

Otherwise you can install it via `go get`

```
go get github.com/zaniluca/pgs3
```

## Configuration

The CLI requires some variables to be defined in order to access the PostgreSQL database and the S3 bucket. These can be passed as environment variables or as flags when running the CLI. They can be passed in both `restore` and `backup` commands.

| Flag                          | Description                                                                    | Env                          | Required |
| ----------------------------- | ------------------------------------------------------------------------------ | ---------------------------- | -------- |
| `--postgres-db` or `-d`       | PostgreSQL Database name                                                       | `PGS3_POSTGRES_DB`           | Yes      |
| `--postgres-user` or `-U`     | PostgreSQL user                                                                | `PGS3_POSTGRES_USER`         | Yes      |
| `--postgres-password` or `-P` | PostgreSQL password                                                            | `PGS3_POSTGRES_PASSWORD`     | Yes      |
| `--postgres-host` or `-H`     | PostgreSQL host                                                                | `PGS3_POSTGRES_HOST`         | No       |
| `--postgres-port` or `-p`     | PostgreSQL port                                                                | `PGS3_POSTGRES_PORT`         | No       |
| `--aws-access-key-id`         | AWS Key ID with permission to Read and Write the specific S3 Bucket (at least) | `PGS3_AWS_ACCESS_KEY_ID`     | Yes      |
| `--aws-secret-access-key`     | AWS Secret Key, comes with the AWS Key ID                                      | `PGS3_AWS_SECRET_ACCESS_KEY` | Yes      |
| `--aws-region` or `-r`        | AWS Region the bucket is stored                                                | `PGS3_AWS_REGION`            | Yes      |
| `--s3-bucket` or `-b`         | AWS S3 Bucket name (must be accessible with provided access key)               | `PGS3_AWS_S3_BUCKET`         | Yes      |
| `--s3-endpoint`               | AWS S3 custom endpoint if using a custom S3 compatible provider                | `PGS3_AWS_S3_ENDPOINT`       | No       |

## `backup` command

The `backup` command is used to backup a PostgreSQL database to an S3 bucket. It uses `pg_dump` to create a dump of the database and then uploads it to the specified S3 bucket.

```sh
pgs3 backup [flags]
```

Available flags are:

| Flag                   | Description                                                                                                                      | Env    | Required |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `--schedule`           | Specify a schedule in [crontab](https://it.wikipedia.org/wiki/Crontab) _(es: 1 \* \* \* \*)_ format for when to execute a backup | _None_ | No       |
| `--restore-on-startup` | Before starting the backup, restore the database from the latest backup in S3 (requires `--schedule`)                            | _None_ | No       |
| `--keep-days`          | Number of days to keep backups in S3                                                                                             | _None_ | No       |
| `--pgdump-extra-opts`  | Extra options to pass to `pg_dump`                                                                                               | _None_ | No       |

## `restore` command

Restore a backup from the configured bucket in S3

```sh
pgs3 restore [flags]
```

Available flags are:

| Flag       | Description                               | Env    | Required |
| ---------- | ----------------------------------------- | ------ | -------- |
| `--latest` | Restore from the most recent backup in S3 | _None_ | Yes      |

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
