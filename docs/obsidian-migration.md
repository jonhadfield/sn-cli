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
to them from a page that gives them context. Every layout writes a `Home.md`
linking to the rest.

### `--moc-style`

**`flat`** (the default) builds MOCs two ways at once: your most-used
top-level tags each get one, and the content of your notes is analysed for
recurring themes, so groupings you never tagged get one too.

**`hierarchical`** mirrors your Standard Notes tag tree. Each tag's MOC links
to the MOCs of its child tags as well as to its own notes, and `Home.md` lists
only the tags that have no parent. If none of your tags are nested there is no
tree to follow, so this falls back to `flat`.

**`topic`** ignores tags and builds MOCs only from the themes found by content
analysis. This suits a vault whose notes are barely tagged. Themes with fewer
notes than `MinNotesPerMOC` are left out, and if nothing is found, `Home.md`
says so rather than being silently empty.

**`para`** sorts your tags into the four
[PARA](https://fortelabs.com/blog/para/) categories — Projects, Areas,
Resources and Archive. The sorting is a guess based on the tag's name
(`sprint-14` reads as a project, `archived-2024` as archive), and anything that
matches nothing lands in Resources. Each MOC says as much at the top, so treat
it as a starting point to rearrange rather than a judgement about your notes.

**`auto`** picks for you: `hierarchical` if any of your tags are nested,
`topic` if you have at least 20 notes and few tags relative to that, otherwise
`flat`.

### `--moc-depth`

How many levels of MOC `hierarchical` creates, defaulting to 2. Tags below the
limit do not get their own MOC; their notes are listed on the deepest MOC
above them, under a "Notes from sub-categories" heading, so no note becomes
unreachable. Depth must be between 1 and 10. The other layouts ignore it.

## Afterwards

Point Obsidian at the output directory with *Open folder as vault*. The
wikilinks and frontmatter tags are picked up with no further configuration.

If you want the notes somewhere other than Obsidian, `sn export` writes
Markdown, HTML or JSON, and has layouts for Hugo and Jekyll. See the
[README](../README.md#export).
