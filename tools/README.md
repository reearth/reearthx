# `github.com/reearth/reearthx/tools`

## i18n-extract

Extract all `i18n.T("hoge")`, `i18n.Message{...}`, and `i18n.LocalizeConfig{...}` from source codes, and write new messages to YAML files

- `i18n` package should be imported from `github.com/reearth/reearthx/i18n`. (renaming is OK)
- Existing translations are retained, but translations that are no longer used are deleted.

```
go run github.com/reearth/reearthx/tools i18n-extract -l en,ja --outdir i18n .
```

for

```go
package example

import "github.com/reearth/reearthx/i18n"

var Hello = i18n.T("hello")
```

results in the following data being written to `en.yml` and `ja.yml`:

```yml
hello:
```

If you want to check for forgotten translations on CI, `--fail-on-update` is helpful. It returns an invalid exit code when there is any update to translation files:

```
go run github.com/reearth/reearthx/tools i18n-extract -l en,ja --outdir i18n --fail-on-update .
```
