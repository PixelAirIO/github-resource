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
    kind: prs # One of: prs, pr
    access_token: gh_pat...
    repository: owner/repo
    # See below for config options, depending on which kind is selected
```

The `kind`, `access_token`, `repository` fields are always required. The
following sections explain each field's usage.

## `kind` - Picking the Resource to Track

The following `kind`'s are supported:

* [`prs`](#kind-prs) - Returns a list of pull requests.
* [`pr`](#kind-pr) - Work with a single Pull Request, changing check statuses or posting comments.

## `access_token` - Configuring Authentication

Authentication is optional if you're accessing public repositories, but you'll
likely want to configure it to avoid rate-limits. Create a [Personal Access
token](https://github.com/settings/personal-access-tokens) (classic or
fine-grained work).

You can also configure the the resource to use a token from a GitHub or OAuth
app. See the GitHub docs for details:
https://docs.github.com/en/graphql/guides/forming-calls-with-graphql#authenticating-with-graphql

## `repository`

This is the repository to target in the format of `OWNER/REPO`. For example, the
repository `https://github.com/example/my-app` would become `repository:
example/my-app`.


## Custom Endpoints

The endpoints used can be configured by setting the following:

- `graphql_endpoint` - Defaults to the GraphQL endpoint `https://api.github.com/graphql`
- `rest_endpoint` - Defaults to the REST endpoint `https://api.github.com`
- `host_endpoint` - Defaults to `https://github.com`, where repositories are hosted
- `skip_ssl_verification` - Skips SSL verification

We try to mostly use the GraphQL API because you're less likely to hit API rate
limits compared to the REST API.

The following table outlines the required permissions for each `kind`.

<table>
    <tr>
        <th><code>kind</code></th>
        <th>Classic Token</th>
        <th>Fine-grained Token</th>
    </tr>
    <tr>
        <td><code>prs</code></td>
        <td><code>repo:public_repo</code></td>
        <td>Public repository access or Repository permission "Read access to metadata"</td>
    </tr>
    <tr>
        <td><code>pr</code></td>
        <td><code>repo:status</code></td>
        <td>Repository permission "Commit statuses" (Write)</td>
    </tr>
</table>

---

The next sections describe how to configure each `kind` of this resource type.

---

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
        <td>
            <code>states</code><em>(Optional)<em>
            <br>
            Default Value: <code>["OPEN"]</code>
        </td>
        <td>A list of PR statuses to filter PR's by. Allowed values are: <code>OPEN</code>, <code>CLOSED</code>, <code>MERGED</code>.</td>
    </tr>
    <tr>
        <td><code>labels</code><em>(Optional)<em></td>
        <td>A list of label(s) to filter PR's by. A PR containing any of the labels will be matched.</td>
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

Allows you to interact with a single Pull Request. Will track commits pushed to
the pull request and allow you to update the status checks of the PR and leave
comments.

`source` has the following additional fields:
<table>
    <tr>
        <th>Field Name</th>
        <th>Description</th>
    </tr>
    <tr>
        <td>
            <code>number</code><em>(Required)<em>
        </td>
        <td>The PR number that the resource will interact with.</td>
    </tr>
    <tr>
        <td>
            <code>merge_strategy</code><em>(Optional)<em>
            <br>
            <em>Defaults to <code>merge</code></em>
        </td>
        <td>Dictates how the PR will be checked out. Can be one of:
            <ul>
                <li><code>merge</code> - Will checkout the branch the PR wants to merge into and locally merge the PR into that branch.</li>
                <li><code>rebase</code> - Will checkout the branch the PR wants to merge into and locally rebase the PR on the latest commit of that branch.</li>
                <li><code>checkout</code> - Only checks out the PR branch.</li>
            </ul>
        </td>
    </tr>
    <tr>
        <td>
            <code>depth</code><em>(Optional)<em>
        </td>
        <td>Shallow clone the repository using git's <code>--depth</code> flag</td>
    </tr>
    <tr>
        <td>
            <code>submodules</code><em>(Optional)<em>
        </td>
        <td>Set to <code>true</code> if you want submodules to be cloned.</td>
    </tr>
    <tr>
        <td>
            <code>fetch_tags</code><em>(Optional)<em>
        </td>
        <td>Set to <code>true</code> if you want tags to be fetched.</td>
    </tr>
    <tr>
        <td>
            <code>disable_git_lfs</code><em>(Optional)<em>
        </td>
        <td>Set to <code>true</code> to not download LFS files.</td>
    </tr>
</table>

The `get` checks out a commit from the Pull Request and locally merges them into
the target branch (unless `merge_strategy: checkout` is used).

The `put` step can set the status on a commit of the Pull Request. One instance
of the resource can be used to set multiple statuses on the PR by calling `put`
with different `params`.

The `put` step has the following params:

<table>
    <tr>
        <th>Field Name</th>
        <th>Description</th>
    </tr>
    <tr>
        <td>
            <code>ref</code><em>(Required)<em>
        </td>
        <td>The commit SHA that the PR check will be matched with on GitHub.</td>
    </tr>
    <tr>
        <td>
            <code>name</code><em>(Required)<em>
        </td>
        <td>The name of the check that will be displayed in the list of PR
        checks for the PR (e.g. `unit-tests`, `integration`). You can use <a
        href="https://concourse-ci.org/docs/resource-types/implementing/#build-metadata">Build
        Metadata</a> in the name like <code>$BUILD_JOB_NAME</code>.</td>
    </tr>
    <tr>
        <td><code>status</code><em>(Required)<em></td>
        <td>One of: `pending`, `success`, `error`, or `failure`</td>
    </tr>
    <tr>
        <td><code>description</code><em>(Optional)<em></td>
        <td>Description that will appear alongside the <code>name</code> of the PR check.</td>
    </tr>
</table>
