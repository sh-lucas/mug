# ☕ Take a coffee

__Mug__ is a code-rebuilding tool, environment variable wrapper and router-glue-code-generating cli-tool written in go.   
This is meant to make the golang developer's life a bit easier with batteries-included experience.    
It loads your environment and generates a folder `cup` in the current folder.   

Today, `mug` is currently generating two packages inside of this folder:
- __cup_router__: a simple package that routes every `//mug:handler <path>` to the default http's handler. Use `cup_router.Route()` to bind to the port.
- __cup__: the main generated package; this includes all the current environment variables in your local `.env` file; they are securely loaded and injected by mug.   

Also, typing `mug` in the terminal will work like `air`, the auto-rebuild tool for go.    
The advantage is that the .env file is automatically loaded and injected in the process.    


## Installing

```bash
go install github.com/sh-lucas/mug/cmd/mug@latest
```


## Why "mug"?

The name?   
> Because it's a cool name.

Currently it's still just a toy project, but it works.  


## TODOS:
- Auto-generating .env.example file.
- Export default Debounce, Throttle, and stuff.
- Export a default Rabbit, Mongo and SQL wrappers.
- Using httpRouter, faster json, and dependency injection.
- Maybe running docker composes in the way with mug.  
- Test suites.    
- `mug build` for cool dockerization.   