# Phloem Acceptance Tests

## Running locally

Default runs exclude WIP scenarios and the brew-release gate:

```bash
go test ./phloem/test/acceptance -v
```

### Brew-release gate

The brew-release acceptance gate is tagged as `@brew_gate` and is **disabled by default**.
Enable it explicitly with `GODOG_TAGS`:

```bash
GODOG_TAGS=@brew_gate go test ./phloem/test/acceptance -v
```

You can combine tags as needed, for example:

```bash
GODOG_TAGS='@brew_gate&&@smoke' go test ./phloem/test/acceptance -v
```

### WIP scenarios

Scenarios tagged `@wip` are excluded from runs by default. Use `GODOG_TAGS` to opt in if needed.
