# Gateway backed by a CAR File

This is an example that shows how to build a Gateway backed by the contents of
a CAR file. A [CAR file](https://ipld.io/specs/transport/car/) is a Content
Addressable aRchive that contains blocks.

## Build

```bash
> go build -o gateway
```

## Usage

First of all, you will need some content stored as a CAR file. You can easily
export your favorite website, or content, using:

```
ipfs dag export <CID> > data.car
```

Then, you can start the gateway with:


```
./gateway -c data.car -p 8040
```

Now you can access the gateway in [127.0.0.1:8040](http://127.0.0.1:8040). It will
behave like a regular IPFS Gateway, except for the fact that all contents are provided
from the CAR file. Therefore, things such as IPNS resolution and fetching contents
from nodes in the IPFS network won't work.