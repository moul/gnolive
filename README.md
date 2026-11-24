# gnolive

Watch a local Gno source file and print each changed, parseable revision,
with comments removed. Originally prototyped in
[gnolang/gno#2617](https://github.com/gnolang/gno/pull/2617).

```sh
go install github.com/moul/gnolive@latest
gnolive -file example.gno
gnolive -once -file example.gno
```

The watcher follows edits and atomic saves. Parse errors go to stderr so
you can keep editing. Each snapshot is written to stdout. The tool uses
Go's parser for syntax and never evaluates a realm or submits a transaction.

See [LICENSE.md](LICENSE.md).
