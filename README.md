# pkg - build linux packages

pkg is a small platform-independent tool for building Linux
.deb or .rpm packages. Notably pkg does not depend on system
utilities like `rpmbuild` or the `dpkg-*` suite of tools.

Most of pkg's magic is provided by [NFPM][nfpm], pkg adds a
command line wrapper and also acts as a
[Github Action](#github-action).

## Usage:

```
Usage:
  pkg [OPTIONS] [package.yaml]

Application Options:
      --name=         package name
      --version=      package version
      --arch=         package arch
      --deb           build a .deb package
      --rpm           build a .rpm package

Help Options:
  -h, --help          Show this help message
```

Example:

```
pkg --name my-package --version 1.0.0 --arch x86_64 --deb package.yaml
```

## package.yaml

The package spec file contains metadata that is identical
for every package.

```yaml
meta:
  description: My Package
  vendor: Me
  maintainer: Me
files:
  "/usr/bin/my-package":
    file: my-package
    mode: "0755"
    user: "root"
  "/etc/my-package/config":
    file: etc/config
    mode: "0644"
    user: "foo"
    keep: true
units:
  - etc/systemd/system/my-package.service
scripts:
  "post-install": scripts/post-inst
user: foo
```

# Github Action

This repository is also a [GitHub Action][action].

## Inputs

| Name          | Description                     |
| ------------- | ------------------------------- |
| *name*        | Package name                    |
| *version*     | Package version                 |
| *arch*        | Package architecture            |
| *format*      | Package format                  |
| *package*     | Package spec file               |

## Outputs

| Name          | Description                     |
| ------------- | ------------------------------- |
| *package*     | Package file name               |

## Usage

```yaml
      - uses: kentik/pkg@v1
        with:
          name: my-package
          version: 0.0.1
          arch: x86_64
          format: rpm
          package: package.yaml
```

| Name          | Valid Values                    |
| ------------- | ------------------------------- |
| *arch*        | x86_64, aarch64, armv7          |
| *format*      | deb, rpm                        |


## License

   Copyright 2020 Kentik

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.


[nfpm]: https://github.com/kentik/nfpm
[action]: https://github.com/features/actions
