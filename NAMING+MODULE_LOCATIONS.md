# Main module function name
```go
func Module() fx.Option {}
```


# Helpers for group provider function name name
```go 
func {Type}Provider() {Type} {}
```


# Where to put modules
For example we have some module A that implements some http server. 
And this module can be extended with some middlewares or smth similar.
So we should put this module in the directory `http/pnpsomelib`
All extensions should be put in the directory `http/pnpsomelib{extension_name}`
So for example logging middleware should be put in the directory `http/pnpsomeliblogging`
