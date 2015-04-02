# git-anchor

git-anchor anchors the versions of your git dependencies.
Given a deps.json file containing your dependency specification git-anchor deduces the correct versions of these repos and generates a build script which clones the anchored versions.

## Problem

Given a big repo with many git subtrees, you would like to have a build 
script on a continuous integration server for each of the small repos.

To accomplish this, each repo's build script could git clone its 
dependencies and run some tests.  A problem could arise when someone pushes to a repo and the build script gets the newest version of each of its dependencies. The dependent code might not be compatible with the latest version of all its dependencies yet.  This could cause an error if, for example a dependency introduced a breaking change in its API.

A possible solution is that developers need to update git hashes in a 
script. However, since such manual build step requirements are error prone, 
git-anchor comes to the rescue with an automatic solution.

git-anchor scans your git subtrees and finds the newest versions that is currently available and in your big repo.  Using this information it generates a script that can clone and checkout the correct version.  It does a little bit more than that, so maybe an example would better explain the details.

## Example

Say we have a repo, lib.git, which is a git subtree ./src/vt/lib inside a big git repo /workspace/.
The developer creates the following file /workspace/src/vt/lib/deps.json with this content.

```
{
    "Dir": "src/vt/lib",
    "Deps": [
        {
            "Repo": "github.com/gogo/protobuf",
            "Dir": "src/github.com/gogo/protobuf"
        },
        {
            "Repo": "github.com/golang/crypto",
            "Dir": "src/golang.org/x/crypto"
        }
    ]
}
```

Next the developer runs:

```
go run /location/of/git-anchor/git-anchor.go -lang=bash ./src/vt/lib/deps.json > ./src/vt/lib/deps.sh && chmod +x ./src/vt/lib/deps.sh
```

The generation language is specified, currently only bash is supported.
The tests are not language specific so it should be very easy to add your favourite build script language.

The deps.sh script now contains the following logic.

```
if src/vt/lib does not exists then
    error "this script must be run from the root of the coding directory"

# code for src/github.com/gogo/protobuf

if src/github.com/gogo/protobuf does not exist then
    if this is inside a git working tree then
        error "please use git subtree to add this repo to your big repo"
    else
        mkdir -p src/github.com/gogo/protobuf
        git clone github.com/gogo/protobuf src/github.com/gogo/protobuf
        (cd src/github.com/gogo/protobuf && git checkout bc946d07d1016848dfd2507f90f0859c9471681e )
else
    if src/github.com/gogo/protobuf is a git repo then
        if git revision == bc946d07d1016848dfd2507f90f0859c9471681e then
            echo "src/github.com/gogo/protobuf is correct revision"
        else
            echo "WARNING src/github.com/gogo/protobuf has wrong revision"
    else
        if src/github.com/gogo/protobuf is a git subtree then
            if subtree revision == bc946d07d1016848dfd2507f90f0859c9471681e then
                echo "src/github.com/gogo/protobuf is correct revision"
            else
                echo "WARNING src/github.com/gogo/protobuf (please ignore this if you are not using git subtree squash) is wrong revision"
        else
            echo "WARNING src/github.com/gogo/protobuf is a folder which is not a git subtree or a git repo"

# code for src/golang.org/x/crypto

...

```

## Practical Application

When using git subtree we have scripts in our big repo that tell us exactly how to push so as to avoid small mistakes in paths and keep our specific usage of squash for the specific repo consistent.
In the push script I now include the script generation code.

Here is an example.

```
#!/bin/bash
set -xe

go run src/github.com/awalterschulze/git-anchor/git-anchor.go -lang=bash src/vt/lib/deps.json > src/vt/lib/deps.sh
echo "[SUBTREE] src/vt/lib/deps.sh has been regenerated, if there is a diff this should be committed and pushed before doing a subtree push"
git diff --exit-code src/vt/lib/deps.sh

git subtree pull --prefix=src/vt/lib \
    http://.../scm/.../lib.git \
    master

git subtree push --prefix=src/vt/lib \
    http://.../scm/.../lib.git \
    master

git subtree pull --prefix=src/vt/lib \
    http://.../scm/.../lib.git \
    master

```


In the case where there were any changes the pusher now has to make the commit and push, before running the push script again.
This forces the update of the deps.sh script.

## Things that can still go wrong

If you make changes to a dependency that is not pushed upstream the build might break.  This is good thing since it reminds you to push your dependencies as well.

## Future

  - Add support for more languages
  - Extend beyond git subtree by adding support for a folder containing multiple repos.

