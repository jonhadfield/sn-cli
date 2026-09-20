# Migrating to Obsidian

`sn migrate obsidian` writes your notes out as an [Obsidian](https://obsidian.md)
vault: one Markdown file per note, tags in YAML frontmatter, and a set of Maps
of Content linking it together.

It only reads from your Standard Notes account. Nothing is modified or deleted
there.

## Usage

```bash
sn migrate obsidian --output ./my-vault
```

`migrate` has one provider today, `obsidian`, which is also available as `obs`.

### Flags

<!-- markdownlint-disable MD013 -->

| Flag | Default | Description |
|------|---------|-------------|
| `--output`, `-o` | required | Directory to write the vault into |
| `--moc`, `-m` | `true` | Generate Maps of Content |
| `--moc-style` | `flat` | MOC layout: `flat`, `hierarchical`, `para`, `topic`, `auto` |
| `--moc-depth` | `2` | Maximum MOC hierarchy depth |
| `--tag-filter` | | Only export notes with these tags (comma-separated) |
| `--dry-run` | | Report what would be written, without writing it |

<!-- markdownlint-enable MD013 -->

Start with a dry run. It reports the file count and the MOCs it would create
without touching the filesystem:

```bash
sn migrate obsidian --output ./my-vault --dry-run
```

Export a subset by tag:

```bash
sn migrate obsidian --output ./my-vault --tag-filter work,projects
```

Skip the MOCs and just get the notes:

```bash
sn migrate obsidian --output ./my-vault --moc=false
```

## What the output looks like

```text
my-vault/
├── Home.md              # Entry point, links to every MOC below
├── Work MOC.md          # One per top-level tag
├── Projects MOC.md
├── Meetings MOC.md      # One per discovered content theme
└── ...                  # One file per note
```

`Home.md` is written last so it can link to every other MOC. Each `<name> MOC.md`
lists the notes belonging to that tag or theme as wikilinks.

## Note format

Each note becomes a Markdown file beginning with YAML frontmatter:

```markdown
---
title: "Quarterly planning"
tags: [work, planning]
created: 2026-01-14T09:12:04Z
updated: 2026-02-02T16:40:11Z
uuid: 7c1f...
source: standard-notes
---

# Quarterly planning

...
```

The `uuid` and `source` fields make it possible to tell which notes came from
Standard Notes, and to match them back up if you migrate again.

## Maps of Content

A Map of Content is an index note: rather than foldering your notes, you link
to them from a page that gives them context. `sn` builds these two ways at
once:

- **By tag.** Your most-used top-level tags each get a MOC.
- **By theme.** The content of your notes is analysed for recurring themes, and
  any theme covering two or more notes gets a MOC as well. This picks up
  groupings you never tagged.

### A note on `--moc-style` and `--moc-depth`

`--moc-style` accepts five values, and they all currently produce the flat
layout described above: `hierarchical`, `para` and `topic` are declared but not
yet implemented, and `auto` resolves to flat deliberately, because the flat
generator already runs the content analysis. `--moc-depth` is accepted and
carried through the config, but nothing reads it yet.

They are documented here because the flags exist and are validated — passing an
invalid style is an error — not because they change the result today. Use the
defaults until that changes.

## Afterwards

Point Obsidian at the output directory with *Open folder as vault*. The
wikilinks and frontmatter tags are picked up with no further configuration.

If you want the notes somewhere other than Obsidian, `sn export` writes
Markdown, HTML or JSON, and has layouts for Hugo and Jekyll. See the
[README](../README.md#export).
