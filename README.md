# vertcoin-openassets
An [OpenAssets](https://github.com/OpenAssets/open-assets-protocol) implementation for Vertcoin

WARNING: This project is in early development stages. Please refrain from using this for storing anything that has actual value.

# Running on testnet

Firstly, make sure you have a running Vertcoin node on testnet. You can run it by executing `vertcoin-qt` with the `-testnet` parameter. Make sure RPC is enabled by including the following settings in your `vertcoin.conf`:

```
server=1
rpcuser=youruserhere
rpcpassword=yourpasswordhere
```

Once testnet has fully synced, you can download the binary release of Vertcoin OpenAssets, and run it. The first time you start it up it will generate a private key and default config in the folder you start it up. It will probably fail to connect to Vertcoin Core on that first time, since the RPC credentials you configured in `vertcoin.conf` need to be matched in `vertcoin-openassets.conf`. Edit the latter file to match the configured credentials in `vertcoin.conf` and restart Vertcoin OpenAssets. The software should now open its browser-based front-end and allow you to receive (testnet) Vertcoin and Assets.