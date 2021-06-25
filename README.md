# Polyhedron

An example application that sends metrics & traces to Honeycomb

* See golang server code in [server/main.go](./server/main.go).
* See OpenTelemetry Collector configuration in [./config/otelcol-config.yaml](./config/otelcol-config.yaml)

## Running this example locally

1. Install dependencies. On MacOS:

```
brew install go ansible entr
brew install --cask multipass
```

2. Create a `provision/cloud-init.yaml` file with a valid SSH public key in it so you can connect to your servers:

```yaml
ssh_authorized_keys:
  - ssh-rsa abcabc_my_ssh_key my@emailaddress.biz
package_update: true
```

3. Create local virtual machines:

```
./provision/provision.bash
```

Hacky note: the scripts from this point forward assume that it's possible to connect to local vms as `vm1`, `vm2`, and `lb`.
To get everything working you'll need to create aliases from those names to the VM IP addresses in your `~/.ssh/config` file.
You can find those IP addresses by running `multipass list`.

4. Deploy to your hosts (you'll need a Honeycomb API key for this)

```
HNY_API_KEY=????? ./deploy.bash
```

5. Server should be running on the `lb` host at port 80

## Load generator

```
go run client/main.go
```

## Got Docker?

Set HNY_API_KEY and HNY_DATASET_NAME environment variables.

```
docker compose up
```

If you've got [Tilt](https://tilt.dev/), `tilt up` will work, too!
