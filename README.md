# 17vpn

![image](https://user-images.githubusercontent.com/91862792/172811759-851153ee-8e76-4e77-a45a-a11504dce767.png)


### Pre-Installation

Follow the [confluence](https://17media.atlassian.net/wiki/spaces/H/pages/1027244286/OKTA+Pritunl+VPN) install Pritunl client and import profiles first

### Installation

```shell
# install it to your $GOPATH/bin
go install github.com/DavidwfLiao/17vpn@v1.3.0
```

### Usage

```shell
# Initial your OTP key and Pin (first time)
# Enter ID or Server to connect/disconnect
$ 17vpn

# Disconnect all connections
$ 17vpn d

# Connect directly by ID or Server
$ 17vpn 2
```

Switching servers disconnects the current one first. A spinner shows the
daemon's status and elapsed time while it works; each step ends with a result line:

```
✔ Disconnected PREPROD 0.8s
✔ Connected PROD 7.6s
```
