# GitHub Resource

A [Concourse](https://concourse-ci.org/) resource for working with GitHub. This resource can be configured to track and update different resources within GitHub. It's like multiple Concourse resources bundled into one.

Use it in your pipeline:

```yaml
resource_types:
- name: github
  type: registry-image
  source:
    repository: ghcr.io/pixel-air/github-resource
    tag: latest

resources:
- name: prs
  type: github
  source:
    kind: prs # One of: prs, pr, release, repositories
    access_token: gh_pat...
    config:
      # See below for config options, depending on which kind is selected
```

## Picking the Resource `kind`

The following `kind`'s are supported:

* [`prs`](#kind-prs) - Work with multiple Pull Requests at once.
* [`pr`](#kind-pr) - Work with a single Pull Request.
* [`release`](#kind-release) - Track and publish GitHub releases.
* [`repositories`](#kind-repositories) - Lists repositories for a GitHub organization or team. Does not clone the repositories.

## Configuring Authentication

Authentication is optional if you're accessing public repositories, but you'll
likely want to configure it to avoid rate-limits. Create a [Personal Access
token](https://github.com/settings/personal-access-tokens) (classic or
fine-grained is fine). You provide the resource the access token via the
`access_token` field, at the same level as the `kind` field.

The following table outlines the required permissions for each `kind`.

<table>
    <tr>
        <th><code>kind</code></th>
        <th>Classic Token</th>
        <th>Fine-grained Token</th>
    </tr>
    <tr>
        <td><code>prs</code></td>
        <td></td>
        <td></td>
    </tr>
    <tr>
        <td><code>pr</code></td>
        <td></td>
        <td></td>
    </tr>
    <tr>
        <td><code>release</code></td>
        <td></td>
        <td></td>
    </tr>
    <tr>
        <td><code>repositories</code></td>
        <td></td>
        <td></td>
    </tr>
</table>


## `kind: prs`

Returns a list of Pull Requests against a given repository.

Only the `get` step is supported. The `put` step is a no-op and will error if you try to use it.

## `kind: pr`

Allows you to interact with a single Pull Request.

The `get` step tracks commits from the Pull Request and locally merges them into
the target branch.

The `put` step can set the status on a commit of the Pull Request. One instance
of the resource can be used to set multiple statuses on the PR.

## `kind: release`

Tracks and publishes GitHub releases.

## `kind: repositories`

Lists all repositories for a given GitHub organization or team. Does not clone
the repositories. The `put` step is not implemented and will error if you try
to use it.

## `kind: branches`

Lists all branches for a given GitHub repository. The `put` step is not
implemented and will error if you try to use it.
