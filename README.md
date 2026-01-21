<div align="center">
 <img src="https://github.com/iliafrenkel/go-pb/blob/e75e6b12af39d83c527676debcd5b4de9d9a01e1/src/web/assets/bighead.svg" width="128px" height="128px" alt="Go PB, pastebin alternative"/>
 <h1>Go PB - Pastebin alternative, written in Go</h1>

[![License MIT](https://img.shields.io/badge/license-MIT-green)](./LICENSE.txt)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.0-4baaaa.svg)](./docs/CODE_OF_CONDUCT.md)

![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/iliafrenkel/go-pb?include_prereleases&sort=semver)
[![Coverage Status](https://coveralls.io/repos/github/iliafrenkel/go-pb/badge.svg?branch=main)](https://coveralls.io/github/iliafrenkel/go-pb?branch=main)
[![Test](https://github.com/iliafrenkel/go-pb/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/iliafrenkel/go-pb/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/iliafrenkel/go-pb)](https://goreportcard.com/report/github.com/iliafrenkel/go-pb)

</div>

Go PB is paste service similar to [Pastebin](https://pastebin.com) that you can
host yourself. All it does is it allows you to share snippets of text with
others. You paste your text, press the "Paste" button and get a short URL that
you can share with anybody. This is the gist. But there is more!

⚠**Warning**: this project is very much a work in progress. A lot of changes are
made regularly, including breaking changes. This is not a usable product yet!

<div align="center">

  ![Contains Technical Debt](https://github.com/iliafrenkel/go-pb/blob/5c415d61c48a9fe3420d1b32752a32152bc51848/docs/contains%20technical%20debt.png)

</div>

## Features

- ✔ Share text snippets.
- ✔ Syntax highlighting for over 250 languages.
- ✔ Burner pastes - paste will be deleted after the first read.
- ⏳ Set expiration time on a paste.
- ✔ Password protection.
- ✔ Paste anonymously, no need to login.
- ✔ Register and you will be able to see the list of pastes you created.
- ✔ Create private pastes. Once logged in, you can create pastes that no one can see.
- ⏳ Public API to create pastes from command line and 3rd party applications.
- ⏳ Admin interface to manage users, pastes and other settings.

✔ - already implemented,
⏳ - work in progress

---

You can see the progress in our [Roadmap](https://github.com/iliafrenkel/go-pb/projects/1).
If you'd like to contribute, please have a look at the [contribution guide](https://github.com/iliafrenkel/go-pb/blob/4d827459e11965778f8608b97936576bd81b55f6/docs/CONTRIBUTING.md).

As always, if you want to learn together, ask a question, offer help or get in
touch for any other reason, please don't hesitate to participate in
[Discussions](https://github.com/iliafrenkel/go-pb/discussions) or to contact
me directly at [frenkel.ilia@gmail.com](mailto:frenkel.ilia@gmail.com).

---

## Running

To run the application, you have two options - download and run the executable or
pull the container image and run it in Docker/Podman/Kubernetes (see [releases](https://github.com/iliafrenkel/go-pb/releases)).

Once you decided how do you want to run, you need to pick a storage backend. At the moment, the
options are: memory, disk, Postgres database. I'll explain the difference.

### Memory

This backend is intended for local development and debugging or if you just want to try out the
application. It requires no special setup. The obvious drawback is that the data is not preserved
after you shutdown the server. To run it all you need is this:

```shell
go-pb
```

That's it! Open your browser at `http://localhost:8080/` and you are good to go. Here is an example
with some command line options you may want to use:

```shell
go-pb \
  --debug \
  --web-brand-name="Super Paste" \
  --web-brand-tagline="Everything is possible, it's only a matter of time"
```

The `--debug` will allow you to use the developer login (a fake auth provider) as well as see more
information in the log.
The `--web-brand-...` options define the text you see in the heading of every page.

### Disk

This is perfect for small setups, where you don't have a lot of users or the use is light. The only
configuration is the directory to save the data in. I have to give one warning here, the data on
disk is not encrypted. Whoever has access to the directort on disk technically can read everything.

```shell
go-pb db-type=disk --disk-data-dir=./mydata
```

The `mydata` directory must exist.


### Postgress database

This backend is intended for real production with lots of users and and heavy usage. For now, I will
assume that you already have an instance of Postgress database up and running.

```shell
go-pb \
  --db-type=postgres \
  --db-connection=host="db-server user=gopb password=gopb dbname=gopb port=5432"
```


Full configruation reference is below in the [Appendix](#Appendix: Configuration Reference).

## Authentication
Describe various authentication options in debug and production mode and how to set them up.

## Full example
Give an example (ro several) of the full setup - database, auth, etc. Maybe one example to run an
executable against existing database. And another example of the full setup (with the database) in
Docker.

## Production considerations
Explain why it is best to put everything behind a TLS terminating proxy.


## Appendix: Configuration Reference

 **Application options:**
 Each option has a reasonable default value which can be overridden by a corresponding environment
 variable. The environment varaible can in turn be overridden by providing command line parameter.

 ```text
      --debug                   [$GOPB_DEBUG]
      --log-file=               Full path to the log file, default is stdout. If
                                run in container, it is best to log to stdout and
                                let the containerastion system handle the logs.
                                Environment variable: $GOPB_LOG_FILE

timeout:
      --timeout-shutdown=       Server graceful shutdown timeout (default: 10s).
                                The maximum time server has to close all
                                connections and shutdown. There is rarely a need
                                to change this.
                                Environment variable: $GOPB_TIMEOUT_SHUTDOWN
      --timeout-http-read=      Duration for reading the entire request (default:
                                15s), This is to protect from very slow connections
                                to keep the connection open.
                                Environment variable: $GOPB_TIMEOUT_HTTP_READ
      --timeout-http-write=     Duration before timing out writes of the response
                                (default: 15s). Same as above, protect from slow
                                connections.
                                Environment variable: $GOPB_TIMEOUT_HTTP_WRITE
      --timeout-http-idle=      Amount of time to wait for the next request on a
                                persistent (keep-alive) connection. If set to 0,
                                read timeout will be used instead (default: 60s).
                                Environment variable: $GOPB_TIMEOUT_HTTP_IDLE

web:
      --web-host=               Hostname part of the Web server address (default:
                                localhost).
                                Environment variable: $GOPB_WEB_HOST
      --web-port=               Port part of the Web server address (default: 8080)
                                Environment variable: $GOPB_WEB_PORT
      --web-log-file=           Full path to the log file, default is stdout.
                                Environment variable: $GOPB_WEB_LOG_FILE
      --web-log-mode=           Log mode, can be 'debug' or 'production' (default:
                                production)
                                Environment variable: $GOPB_WEB_LOG_MODE
      --web-brand-name=         Brand name shown in the header of every page
                                (default: Go PB)
                                Environment variable: $GOPB_WEB_BRAND_NAME
      --web-brand-tagline=      Brand tagline shown below the brand name (default:
                                "A nice and simple pastebin alternative that you can
                                host yourself.")
                                Environment variable: $GOPB_WEB_BRAND_TAGLINE
      --web-assets=             Path to the assets folder (default: ./assets). This
                                is where all the CSS, JavaScripts, and icons are.
                                Environment variable: $GOPB_WEB_ASSETS
      --web-templates=          Path to the templates folder (default: ./templates).
                                This is where the HTML templates are.
                                Environment variable: $GOPB_WEB_TEMPLATES
      --web-bootstrap-theme=    Name of the bootstrap CSS theme to use. One of the 
                                following: flatly, litera, materia, sandstone, yeti
                                or zephyr. (default: original)
                                Environment variable: $GOPB_WEB_BOOTSTRAP_THEME]
      --web-logo=               Name of the logo image file within the assets folder
                                (default: bighead.svg)
                                Environment variable: $GOPB_WEB_LOGO
      --web-max-body-size=      Maximum size for request's body in bytes (default:
                                10240). Be carefull not to set it too high. Bad
                                people can start using your server as a storage.
                                Environment variable: $GOPB_WEB_MAX_BODY_SIZE

db:
      --db-type=                Database type to use for storage. Can be one of:
                                memory - everything is stored in memory, everyting
                                will be gone after server shutdown. Main use for
                                development and debugging.
                                disk - use files on disk to store the data. Data is
                                persisted and survives restarts. Good for low use
                                setups. ⚠ Data on disk is not encrypted! ⚠
                                postgres - use Postgres database for persistence.
                                Best for high use cases.
                                (default: memory)
                                Environment variable: $GOPB_DB_TYPE
      --db-connection=          Database connection string for Postgres type.
                                Ignored for memory or disk storage types.
                                Environment variable: $GOPB_DB_CONNECTION

auth:
      --auth-secret=            Secret used for JWT token generation/verification.
                                Environment variable: $GOPB_AUTH_SECRET
      --auth-token-duration=    JWT token expiration. (default: 5m)
                                Environment variable: $GOPB_AUTH_TOKEN_DURATION
      --auth-cookie-duration=   Cookie expiration. (default: 24h)
                                Environment variable: $GOPB_AUTH_COOKIE_DURATION
      --auth-issuer=            App name used to oauth requests. (default: go-pb)
                                Environment variable: $GOPB_AUTH_ISSUER
      --auth-url=               Callback url for oauth requests. (default:
                                http://localhost:8080)
                                Environment variable: $GOPB_AUTH_URL
      --auth-github-cid=        Github client id used for oauth.
                                Environment variable: $GOPB_AUTH_GITHUB_CID
      --auth-github-csec=       Github client secret used for oauth.
                                Environment variable: $GOPB_AUTH_GITHUB_CSEC
      --auth-google-cid=        Google client id used for oauth.
                                Environment variable: $GOPB_AUTH_GOOGLE_CID
      --auth-google-csec=       Google client secret used for oauth.
                                Environment variable: $GOPB_AUTH_GOOGLE_CSEC

disk:
      --disk-data-dir=          Directory where pastes are stored. The directory
                                must exist. (default: ./data)
                                Environment variable: $GOPB_DISK_DATA_DIR
      --disk-cache-size=        File system storage cache size.
                                Environment variable: $GOPB_DISK_CACHE_SIZE
      --disk-dir-mode=          File mode for new directories.
                                Environment variable: $GOPB_DISK_DIR_MODE
```
