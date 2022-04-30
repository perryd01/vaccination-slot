# Usage
Example:
```
go run ./tools/network_setup/main.go -reuse
```

## ContainerName
```
-cname string
```  
docker ContainerName (default "vacc_slot")
## Host Port
```
-hport uint
```  
host port number (default 8080)

## Reuse
```
-reuse
```  
reuse container if exists (default false)  
**only works if current containers name equals to the existing one's**
