# cinc-supermarket

A Go client for the [Chef Supermarket API](https://docs.chef.io/supermarket/supermarket_api/).

## Install

    go get github.com/tas50/cinc-supermarket

## Usage

Read endpoints are anonymous:

    c, _ := supermarket.NewClient(supermarket.Config{})
    cb, _, _ := c.Cookbooks.Get(context.Background(), "apache2")

Write endpoints (share, delete) use the Chef v1.3 signed-header
protocol. The credentials are the Supermarket username and the RSA
private key registered on that user's Supermarket profile:

    key, _ := supermarket.LoadKeyFile("/home/me/.chef/me.pem")
    c, _ := supermarket.NewClient(supermarket.Config{
        Username: "me",
        Key:      key,
    })
    c.Cookbooks.Share(ctx, "apache2", "Web Servers", tarball)

A `Client` with no credentials refuses write calls with
`ErrUnauthenticatedWrite` without contacting the server.

Version helpers work on the version strings cookbooks return, without a
full semver dependency: `CompareVersions(a, b)` orders two versions,
`LatestVersion(versions)` picks the newest, and `VersionFromURL(u)`
pulls the dotted version out of a `/versions/<v>` URL (Supermarket
encodes the dots as underscores).

## Coverage

| Resource | Endpoint                                              | Method          |
| -------- | ----------------------------------------------------- | --------------- |
| Cookbook | `GET /api/v1/cookbooks`                               | `List`          |
| Cookbook | `GET /api/v1/cookbooks/:name`                         | `Get`           |
| Cookbook | `GET /api/v1/cookbooks/:name/contingent`              | `Contingent`    |
| Cookbook | `GET /api/v1/cookbooks/:name/versions/:v`             | `GetVersion`    |
| Cookbook | `GET /api/v1/cookbooks/:name/versions/:v/download`    | `Download`      |
| Cookbook | `POST /api/v1/cookbooks` (signed, multipart)          | `Share`         |
| Cookbook | `DELETE /api/v1/cookbooks/:name` (signed)             | `Delete`        |
| Cookbook | `DELETE /api/v1/cookbooks/:name/versions/:v` (signed) | `DeleteVersion` |
| Search   | `GET /api/v1/search`                                  | `Cookbooks`     |
| Tools    | `GET /api/v1/tools`                                   | `List`          |
| Tools    | `GET /api/v1/tools/:slug`                             | `Get`           |
| Tools    | `GET /api/v1/tools-search`                            | `Search`        |
| Users    | `GET /api/v1/users/:username`                         | `Get`           |
| Universe | `GET /universe`                                       | `Get`           |
| Universe | `GET /universe` (streaming)                           | `GetStream`     |
| Health   | `GET /api/v1/health`                                  | `Status`        |
| Health   | `GET /api/v1/metrics`                                 | `Metrics`       |

## Errors

Non-2xx responses are wrapped in an `*ErrorResponse` that decodes
Supermarket's `{error_messages, error_code}` envelope. Use
`errors.Is(err, supermarket.ErrNotFound)` (and friends) to check the
class of failure without dealing with the fact that Supermarket
returns 400 with `error_code=NOT_FOUND` rather than 404.

## License

See LICENSE.
