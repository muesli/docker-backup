# Contributing guidelines

## Pull Requests ##

You can use [The GitHub
flow](https://guides.github.com/introduction/flow/), which mostly just
involves creating a separate branch for each patch you're working on.
Using that method helps prevent merge conflicts later. **Note** that
you should never need to work on the master branch, and never merge
your patches into the master branch (See "syncing" below).

Source code patches should only contain changes related to a single
issue. This helps speed up the review and discussion process. However,
if you're helping fix typos and grammar errors in documentation,
multiple changes in one PR is fine. General rule of thumb for
documentation patches on this project is 5 unrelated changes or fewer
to a PR. But if they are only one-word or letter changes, I can be
flexible and more than 5 will still be gratefully accepted for review.

If you're submitting a documentation-only PR, please add `[skip ci]` to
the commit message; it will prevent automatic code testing on
[Travis](https://travis-ci.org/muesli/docker-backup).


## Syncing ##

Periodically, you'll need to sync your repo with the upstream.
GitHub has instructions for doing this

* [Configuring a remote for a fork](https://help.github.com/articles/configuring-a-remote-for-a-fork/)
  * For step 3 on that page, use https://github.com/muesli/docker-backup.git for the URL.
* [Syncing a Fork](https://help.github.com/articles/syncing-a-fork/)
  * On that page, it shows how to merge the **master** branch (steps 4 & 5).
