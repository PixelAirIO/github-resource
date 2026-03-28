# GitHub Resource

A [Concourse](https://concourse-ci.org/) resource for working with GitHub. This
resource can be configured to track and update different resources within
GitHub. It's multiple Concourse resources bundled into one container image.

Use it in your pipeline:

```yaml
resource_types:
- name: github
  type: registry-image
  source:
    repository: docker.io/pixelairio/github-resource
    tag: latest

resources:
- name: prs
  type: github
  source:
    kind: prs # One of: prs, pr, release, repositories, branches
    access_token: gh_pat...
    # See below for config options, depending on which kind is selected
```

## Picking the Resource `kind`

The following `kind`'s are supported:

* [`prs`](#kind-prs) - Returns a list of pull requests.
* [`pr`](#kind-pr) - Work with a single Pull Request.
* [`release`](#kind-release) - Track and publish GitHub releases.

## Configuring Authentication

Authentication is optional if you're accessing public repositories, but you'll
likely want to configure it to avoid rate-limits. Create a [Personal Access
token](https://github.com/settings/personal-access-tokens) (classic or
fine-grained is fine). You provide the resource the access token via the
`access_token` field.

You can also configure the the resource to use a token from a GitHub or OAuth
app. See the GitHub docs for details:
https://docs.github.com/en/graphql/guides/forming-calls-with-graphql#authenticating-with-graphql

## Custom Endpoint

The endpoint can be configured by setting `api_endpoint` field. It will
default to `https://api.github.com/graphql`. Only the GraphQL API is supported
at this time because you're less likely to hit API rate limits compared to the
REST API.

The following table outlines the required permissions for each `kind`.

<table>
    <tr>
        <th><code>kind</code></th>
        <th>Classic Token</th>
        <th>Fine-grained Token</th>
    </tr>
    <tr>
        <td><code>prs</code></td>
        <td><code>public_repo</code></td>
        <td>Public repository access</td>
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
</table>


## `kind: prs`

Returns a list of Pull Requests against a given repository. Can filter by the
PR's status, labels, draft status, and target branch the PR wants to merge into.

`source` has the following additional fields:
<table>
    <tr>
        <th>Field Name</th>
        <th>Description</th>
    </tr>
    <tr>
        <td><code>owner</code><em>(Required)<em></td>
        <td>The username or organization that owns the repository.</td>
    </tr>
    <tr>
        <td><code>repository</code><em>(Required)<em></td>
        <td>The repository to track PR's from.</td>
    </tr>
    <tr>
        <td>
            <code>states</code><em>(Optional)<em>
            <br>
            Default Value: <code>["OPEN"]</code>
        </td>
        <td>A list of PR statuses to filter PR's by. Allowed values are: <code>OPEN</code>, <code>CLOSED</code>, <code>MERGED</code>.</td>
    </tr>
    <tr>
        <td><code>labels</code><em>(Optional)<em></td>
        <td>A list of label(s) to filter PR's by. A PR must have all listed labels to be matched.</td>
    </tr>
    <tr>
        <td><code>target_branch</code><em>(Optional)<em></td>
        <td>Only track PRs that merge into the specified branch.</td>
    </tr>
    <tr>
        <td>
            <code>exclude_drafts</code><em>(Optional)<em>
            <br>
            Default Value: <code>false</code>
        </td>
        <td>Exclude PRs that are currently drafts.</td>
    </tr>
</table>

Only the `get` step is supported. The `put` step is a no-op and will error if
you try to use it. The `get` step will write the list of PR's to a file,
`prs.json`, as a JSON array of the PR numbers and other information. The numbers
will be saved as strings, not integers.

Example of `prs.json`:
```json
[
  {
    "number": "1234",
    "url": "http://...",
    "target_branch": "main"
  }
]
```

In the case when there are no matching PRs, a special `none` version will be
generated. The `get` step will populate `prs.json` with an empty array that can
be passed to the `across` step.

## `kind: pr`

Allows you to interact with a single Pull Request. Will track commits pushed to the pull request.

The `get` step returns a commit from the Pull Request and locally merges them
into the target branch.

The `put` step can set the status on a commit of the Pull Request. One instance
of the resource can be used to set multiple statuses on the PR.

## `kind: release`

Tracks and publishes GitHub releases.
