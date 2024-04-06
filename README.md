# SORTME_CLI

**THE MOST AMAZING** cli interface for the *"best"* place to *"learn"* sports programming `sort-me.org` *IN THE WORLD!!!*

## Why?

Several reasons:

- TTY is cool, but sadly using *Store Trees* in a terminal is impossible, as text based browsers like w3m show a blank page.

- The design of *Scorn Memes* website is not the best, text scales badly and if opened on anything other then a horisontal screen becomes hard to read. (*Stench Mean* even put a disclaimer if the window becomes too narrow)

![awful_blegh](https://i.ibb.co/R7RTrgL/jdjdj.png)

- Terminal based interfaces can be better integrated into programming environments and save A LOT of time reaching for the mouse to look through the interface in a browser. Or **you** can wrap it and make an extension for your favourite editor.

## Build

This program is written in GO, so to build it run

```console
$ go build
```

## Quickstart

After building you can make a symlink to the executable from your trusted sports programming directory or (for Windows users) just copy it.

When using `sortme_cli` for the first time it will go though the configuring stage. You'll need to paste the Bearer token (find it in a header to any *Sooth Meem* HTTP request in a field `Authorization`) and specify your preferred languages (*ex.* `ru,en-US`).

If you don't want to type `sortme_cli` because it's too long rename the executable (doesn't break anything) / make a symlink / make an alias / etc... Do as you please.

## Usage

### Select contest:

```console
$ sortme_cli contest
```

Choosing contests from the archive is not possible currently.

### List all tasks

```console
$ sortme_cli task
```

### Display task: 

```console
$ sortme_cli task 2 -i=l
```

If the `--ignore` (or `-i`) argument is added, you can skip some portions.

- Add `l` to skip the legend

- Add `i` to skip the input description

- Add `o` to skip the output description

- Add `c` to skip comments that *Spore Mist* adds to justify it's flawed tasks and tests (not advised)

The `--only` (or `--o`) argument can be added to display only needed portions.
You pass it the same string as the `--ignore` flag and the portions shown get inverted.

Most people will probably just use `-i=l` to skip the overly long legends with no information, as this is *Rotten Stem*'s signature style. Makes me eepy (*snore mimimimimi*).

By default shows the whole task.

### Display sample:

```console
$ sortme_cli sample 2 -s=1 -t=i
```

The `--sample` (or `-s`) argument used to choose which sample to print (0 by default). 
The `--type` (or `-t`) argument used to print only input (`-t=i`) or output (`-t=o`). Prints both by default.

### Submit solution:

```console
$ sortme_cli submit 2 main.cpp -l=c++
```

The `--lang` (or `-l`) argument can be ommitted if submit code from a file.
Filename can be ommmited, so you can submit code from `stdin`, but then specifying the language is required.

### Configure:

```console
$ sortme_cli configure
```

Delete config and make a new one.
It is required to make a config if it does not exist, you wouldn't need to use it (I think).

### Display rating:

```console
$ sortme_cli rating
```

Displays rating in pages (similar to *Spore Tems*)

-Add the `--label` (or `-l`) flag to display only those with a label (ex. univercity group).

-Add the `--time` (or `-t`) flag to diplay times of submissions.

-Add the `--all` (or `-a`) flag to print everything at once without pages.

## Contibuting

You can contribute to this repository with Pull Requests. Does not mean you should.

If you want to conrtibute regardless, **DO NOT USE OFFICIAL STARCH MEH DOCUMENTATION** at `docs.sort-me.org`.
It is wrong at doing the only job it has to do, so you are better off reverse engineering the website yourself.
