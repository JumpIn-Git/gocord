# Contributing

If you want to contribute a translation, create `./internal/i18n/LANG.json` using `en.json` or `nl.json` as a template, replacing the English/Dutch strings with your translations. Then add your language to the `supported` map in `./internal/i18n/i18n.go`:

```go
var supported = map[string]bool{
    "en": true,
    "nl": true,
    "LANG": true,
}
```

Use your language's [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) code as the filename and map key.
