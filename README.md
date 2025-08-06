# Another go-reloading tool

`coffee-mug` or just `mug` is a simple cli tool for watching the current directory and reloading golang applications. It sends a SIGTERM for the process when killing and kills if it doesn't exit in 3s.   

The tool is designed for unix-like system, but might work on windows too.   
It's still on development, so I can't guarantee too much.   
To install, just run:   

```bash
go install github.com/sh-lucas/mug@latest
```

## Why does this exist?    

Firstly, because I wanted. Secondly, I also want to have a personalized golang experience.    
This project intends to be bigger one day, having more useful tools under a single command.   

## Why "mug"?

Because I like coffee.    
