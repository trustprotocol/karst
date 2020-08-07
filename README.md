# Karst &middot; [![Build Status](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fcrustio%2Fcrust%2Fbadge&style=flat)](https://github.com/crustio/karst/actions?query=workflow%3AGo)
Karst is a storage adapter integrated with FS(file system) and sWorker(storage work inspector) of Crust protocol. It manages storage resources to serve the storage market.

<a href='https://web3.foundation/'><img width='220' alt='Funded by web3 foundation' src='docs/img/web3f_grants_badge.png'></a>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<a href='https://builders.parity.io/'><img width='260' src='docs/img/sbp_grants_badge.png'></a>

## Compiler Environment
```shell
go version >= go1.13.4
```

# Config
Configuration file will be created by running './karst init' in $KARST_PATH/config.json (Default location is $HOME/.karst/config.json). You can change it like:

```json
{
  "base_url": "0.0.0.0:17000",
  "crust": {
    "address": "5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX",
    "backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
    "base_url": "127.0.0.1:56666/api/v1",
    "password": "123456"
  },
  "fastdfs": {
    "max_conns": 100,
    "tracker_addrs": ["172.16.3.15:22122"]
  },
  "log_level": "debug",
  "tee_base_url": "127.0.0.1:12222/api/v0"
}
```

- 'base_url' is karst url
- 'crust.address' is your chain account
- 'crust.backup' is your backup for chain
- 'crust.base_url' is crust api url for chain
- 'crust.password' is password for chain
- 'fastdfs.max_conns' is the maximum number of connections for fastdfs
- 'fastdfs.tracker_addrs' is the addresses of trackers for fastdfs
- 'log_level' can be set as debug mode to show debug information
- 'tee_base_url' is tee base url

## Install & Run

### Install

```shell
sudo ./install.sh # for linux
```

```shell
go build # for mac and windows, then move the kasrt bin to commands folder or add it to PATH
```

### Run

For provider

- Set $KARST_PATH to change karst installation location, default location is $Home/.karst/
```shell
  karst init
``` 
- Change your configurations
```shell
  vim ~/.karst/config.json
```
- Start karst
```shell
  karst daemon
```
- Register your karst external address
```shell
  karst register ws://localhost:17000 1000
```
- List your stored files
```shell
  karst list
```

- Delete your stored files
```shell
  karst delete e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef
```

For client

- Set $KARST_PATH to change karst installation location, default location is $Home/.karst/
```shell
  karst init
```
- Change your configurations
```shell
  vim ~/.karst/config.json
```
- Start karst
```shell
  karst daemon
```
- Split file, and you can send the splited files to fs
```shell
  karst split /home/crust/test/karst/1M.bin /home/crust/test/karst/output
```
- Fill the stored_key of each fragment in fs. Then declare the file to chain and request provider to generate store proof
```shell
  karst declare "{\"hash\":\"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef\",\"size\":1048567,\"links_num\":1,\"stored_key\":\"\",\"links\":[{\"hash\":\"055162be19abb648f4ff47f1292574192d9b7131f900f609bee0dd79c0e60970\",\"size\":1048567,\"links_num\":0,\"stored_key\":\"group1/M00/00/5E/wKgyC17fI0KAYzlEAA__9-56uVA3640992\",\"links\":[]}]}" 1000 5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX
```
- Try to get file stored information
```shell
  karst obtain e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef 5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX
```
- After downloading the file successfully, use 'finish' to help provider to clear file
```shell
  karst finish "{\"hash\":\"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef\",\"size\":1048567,\"links_num\":1,\"links\":[{\"hash\":\"055162be19abb648f4ff47f1292574192d9b7131f900f609bee0dd79c0e60970\",\"size\":1048567,\"links_num\":0,\"links\":[],\"stored_key\":\"group1/M00/00/00/wKgyC17sdDyAYVuQAA__9-56uVA2354372\"}],\"stored_key\":\"\"}" 5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX 
```

## Docker model
Please refer to [karst docker mode](docs/docker.md)

## Interface

Karst provides plenty of getting and controlling interfaces, please refer to [interface](docs/interface.md)

## Contribution

Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to crust, please **fork, fix, commit and send a pull request for the maintainers to review and merge into the main codebase**.

### Rules

Please make sure your contribution adhere to our coding guideliness:

- **No --force pushes** or modifying the master branch history in any way. If you need to rebase, ensure you do it in your own repo.
- Pull requests need to be based on and opened against the `master branch`.
- A pull-request **must not be merged until CI** has finished successfully.
- Make sure your every `commit` is [signed](https://help.github.com/en/github/authenticating-to-github/about-commit-signature-verification)

### Merge process

Merging pull requests once CI is successful:

- A PR needs to be reviewed and approved by project maintainers;
- PRs that break the external API must be tagged with [`breaksapi`](https://github.com/crustio/crust-tee/labels/breakapi);
- No PR should be merged until **all reviews' comments** are addressed.

## License

[GPL v3](LICENSE)
