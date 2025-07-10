# Another go-reloading tool

`coffee-mug` or just `mug` is a simple cli tool for watching the current directory and reloading golang applications. It sends a SIGTERM for the process when killing and kills if it doesn't exit in 3s.   

The tool is designed for unix-like system, but might work on windows too.   
It's still on development, so I can't guarantee too much.   
To install, just run:   

```bash
go install github.com/sh-lucas/mug
```