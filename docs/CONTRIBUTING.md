# Contributing to MOTH
We love your input! We want to make contributing to this project as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features

## We Develop with Github
We use github to host code, to track issues and feature requests, as well as accept pull requests.

## We Use the Feature Branch Workflow

We use the [Feature Branch Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/feature-branch-workflow).

Pull requests are the best way to propose changes to the codebase. 
We actively welcome your pull requests:

1. Fork the repo. Optionally create a branch from `main`.
2. If you've changed code, modify tests that fail on the old code, and pass on the new code.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Update [CHANGELOG.md](CHANGELOG.md)
7. Issue that pull request!

## We Deploy to a Variety of Architectures
MOTH is most often deployed using Docker, but we strive to ensure that it can easily be run outside of a Docker environment. Please ensure that and changes will not break or substantially alter Dockerized deployments and that, conversely, changes will not so substantially tie MOTH to Docker or particular Docker deployment that it becomes impractical to run MOTH anywhere but inside of Docker

## Any contributions you make will be under the MIT Software License
When you submit code changes, your submissions are understood to be under the same [MIT License](http://choosealicense.com/licenses/mit/) that covers the project. Feel free to contact the maintainers if that's a concern.

## Report bugs using Github's [issues](https://github.com/dirtbags/moth/issues)
We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/dirtbags/moth/issues/new); it's that easy!

## Write bug reports with detail, background, and sample code

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific!
  - Give sample code if you can.
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

## Use a Consistent Coding Style

### Go
* Run it through `gofmt`

### Javascript
* We use Javascript ASI

## References
This document was adapted from the open-source contribution guidelines from [https://gist.github.com/briandk/3d2e8b3ec8daf5a27a62]
