<div align="center">
 <img src="https://github.com/iliafrenkel/go-pb/blob/e75e6b12af39d83c527676debcd5b4de9d9a01e1/src/web/assets/bighead.svg" width="128px" height="128px" alt="Go PB, pastebin alternative"/>
 <h1>Go PB - Pastebin alternative, written in Go</h1>

[![License MIT](https://img.shields.io/badge/license-MIT-green)](./LICENSE.txt)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com) 
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.0-4baaaa.svg)](./docs/CODE_OF_CONDUCT.md) 
![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/iliafrenkel/go-pb?include_prereleases&sort=semver)
[![codecov](https://codecov.io/gh/iliafrenkel/go-pb/branch/main/graph/badge.svg?token=WR1DWNVE58)](https://codecov.io/gh/iliafrenkel/go-pb)
![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/iliafrenkel/go-pb/Test/main?label=test)

</div>

Go PB is paste service similar to [Pastebin](https://pastebin.com) that you can
host yourself. All it does is it allows you to share snippets of text with
others. You paste your text, press the "Paste" button and get a short URL that
you can share with anybody. This is the gist. But there is more!

⚠**Warning**: this project is very much a work in progress. A lot of changes are
made regularly, including breaking changes. This is not a usable product yet!

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
