# movistarcli

Unofficial CLI for managing your Movistar HGU router from the terminal. (tested on Askey RTF3505VW).

![](https://i.imgur.com/JAAI9R6.png)

## Installation

```bash
go install github.com/nuriofernandez/movistarcli@latest
```

**Note:** `$GOPATH` must be on your `$PATH` in order to work. ```export PATH=${PATH}:`go env GOPATH`/bin```

## Authentication

The router password is resolved in this order:

1. `--password` flag
2. `MOVISTAR_PASSWORD` environment variable
3. `~/.config/Movistar/credentials` file

To use the credentials file, create it with:

```
password=yourpassword
```

## Commands

### Reboot

```bash
movistarcli reboot
```

### Devices

List all devices connected to the router:

```bash
movistarcli devices
```

### Port forwarding

**List rules:**
```bash
movistarcli ports
movistarcli ports list
```

**Add a rule:**

The `-p` flag format is `internal[/external]`, where `:` separates a port range:

| Flag | int | ext |
|---|---|---|
| `-p 80` | 80 | 80 |
| `-p 80:81` | 80–81 | 80–81 |
| `-p 80/90` | 80 | 90 |
| `-p 80:81/90` | 80–81 | 90–91 (derived) |
| `-p 80:81/90:91` | 80–81 | 90–91 |

```bash
movistarcli ports add --name "HTTP"  --protocol TCP  --address 192.168.1.100 -p 80
movistarcli ports add --name "HTTP"  --protocol TCP  --address 192.168.1.100 -p 80/8080
movistarcli ports add --name "Game"  --protocol UDP  --address 192.168.1.50  -p 8000:8010
movistarcli ports add --name "Game"  --protocol UDP  --address 192.168.1.50  -p 8000:8010/9000:9010
```

**Update a rule** (only the provided flags are changed):
```bash
movistarcli ports update 32 --name "New Name"
movistarcli ports update 32 --address 192.168.1.200
movistarcli ports update 32 -p 80
movistarcli ports update 32 -p 8000:8010/9000
```

**Enable / disable a rule:**
```bash
movistarcli ports enable 32
movistarcli ports disable 32
```

**Delete a rule:**
```bash
movistarcli ports delete 32
```
