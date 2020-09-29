# Interface

## Interface for merchant
### Register /api/v0/cmd/register
#### Input
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456",
	"karst_address": "ws://localhost:17000",
	"storage_price": "1000"
}
```

#### Return
```json
{
	"info":"Register 'ws://127.0.0.1:17000' successful in 18.624281178s ! You can check it on crust.",
	"status":200
}
```

**ps: register needs call chain's rpc, so you may wait for couple seconds waiting for chain's confirm**

### List /api/v0/cmd/list
#### Input(list all files)
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456"
}
```

#### Return(list all files) 
```json
{
	"info":"List all files successfully in 38.361µs !",
	"files":[{"hash":"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef","size":1048567,"sealed_hash":"b6f5755923f5e82ed84274ad5f378d49f32f765a8c6a4a9921046226b5e21e97","sealed_size":1049127}],"status":200
}
```

#### Input(list file details)
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456",
	"file_hash": "e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef"
}
```

#### Return(list file details) 
```json
{
	"info":"List file 'e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef' successfully in 50.668µs !",
	"file":{"merkle_tree":{"hash":"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef","size":1048567,"links_num":1,"links":[{"hash":"055162be19abb648f4ff47f1292574192d9b7131f900f609bee0dd79c0e60970","size":1048567,"links_num":0,"links":[],"stored_key":"group1/M00/00/5E/wKgyC17fI0KAYzlEAA__9-56uVA3640992"}],"stored_key":""},"merkle_tree_sealed":{"hash":"b6f5755923f5e82ed84274ad5f378d49f32f765a8c6a4a9921046226b5e21e97","size":1049127,"links_num":1,"links":[{"hash":"0171c4f38bf451d1ab2250804ec24946f59d064a5d411074c4dc768724cc8d18","size":1049127,"links_num":0,"links":null,"stored_key":"group1/M00/00/5E/wKgyC17hn12AFsmSABACJ8njfU84021465"}],"stored_key":""},
	"status":200
}
```

### Delete /api/v0/cmd/delete
#### Input(delete automatically clear files that are not in the order list)
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456"
}
```

#### Return(delete automatically clear files that are not in the order list)
```json
{
	"info":"Delete '10' files successful in 1.624281178s ! You can check it on crust.",
	"status":200
}
```

#### Input(delete one file)
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456",
	"file_hash": "e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef"
}
```

#### Return(delete one file)
```json
{
	"info":"Delete 'e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef' successful in 1.624281178s ! You can check it on crust.",
	"status":200
}
```

## Websocket interface (for client)
### Split /api/v0/cmd/split
#### Input
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456",
	"file_path": "/home/crust/test/karst/10M.bin",
	"output_path": "/home/crust/test/karst/o"
}
```

**ps: 'file_path' and 'output_path' must be absolute path**

#### Return
```json
{
	"info":"Split '/home/crust/test/karst/1M.bin' successfully in 6.962893ms ! It root hash is 'e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef'.",
	"merkle_tree":"{\"hash\":\"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef\",\"size\":1048567,\"links_num\":1,\"stored_key\":\"\",\"links\":[{\"hash\":\"055162be19abb648f4ff47f1292574192d9b7131f900f609bee0dd79c0e60970\",\"size\":1048567,\"links_num\":0,\"stored_key\":\"\",\"links\":[]}]}",
	"status":200
}
```

### Declare /api/v0/cmd/declare
#### Input
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456",
	"merkle_tree": "{\"hash\":\"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef\",\"size\":1048567,\"links_num\":1,\"stored_key\":\"\",\"links\":[{\"hash\":\"055162be19abb648f4ff47f1292574192d9b7131f900f609bee0dd79c0e60970\",\"size\":1048567,\"links_num\":0,\"stored_key\":\"group1/M00/00/5E/wKgyC17fI0KAYzlEAA__9-56uVA3640992\",\"links\":[]}]}",
	"duration": "1000",
	"merchant": "5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX"
}
```

#### Return
```json
{
	"info":"Declare successfully in 17.616240658s ! Store order hash is '0x4aa1726f451e7f9759edf29a71ad045aab6861362be01f78d89421dc040d4d95'.","store_order_hash":"0x4aa1726f451e7f9759edf29a71ad045aab6861362be01f78d89421dc040d4d95","status":200
}
```

### Obtain /api/v0/cmd/obtain
#### Input
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456",
	"file_hash": "e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef",
	"merchant": "5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX"
}
```

#### Return
```json
{
	"info":"Obtain 'e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef' from '5HZFQohYpN4MVyGjiq8bJhojt9yCVa8rXd4Kt9fmh5gAbQqA' successfully in 33.938813ms !",
	"merkle_tree":"{\"hash\":\"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef\",\"size\":1048567,\"links_num\":1,\"links\":[{\"hash\":\"055162be19abb648f4ff47f1292574192d9b7131f900f609bee0dd79c0e60970\",\"size\":1048567,\"links_num\":0,\"links\":[],\"stored_key\":\"group1/M00/00/00/wKgyC17sdDyAYVuQAA__9-56uVA2354372\"}],\"stored_key\":\"\"}",
	"status":200
}
```

### Finish /api/v0/cmd/finish
#### Input
```json
{
	"backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}",
	"password": "123456",
	"merkle_tree":"{\"hash\":\"e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef\",\"size\":1048567,\"links_num\":1,\"links\":[{\"hash\":\"055162be19abb648f4ff47f1292574192d9b7131f900f609bee0dd79c0e60970\",\"size\":1048567,\"links_num\":0,\"links\":[],\"stored_key\":\"group1/M00/00/00/wKgyC17sdDyAYVuQAA__9-56uVA2354372\"}],\"stored_key\":\"\"}",
	"merchant": "5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX"
}
```

#### Return
```json
{
	"info":"Request merchant 'e2f4b2f31c309e18dbe658d92b81c26bede6015b8da1464b38def2af7d55faef' to finish '5HZFQohYpN4MVyGjiq8bJhojt9yCVa8rXd4Kt9fmh5gAbQqA' successfully in 3.91568ms !",
	"status":200
}
```

## Interface for sWorker
### Node data /api/v0/node/data
#### Send backup message to identity your authority
```json
{
    "backup": "{\"address\":\"5FqazaU79hjpEMiWTWZx81VjsYFst15eBuSBKdQLgQibD7CX\",\"encoded\":\"0xc81537c9442bd1d3f4985531293d88f6d2a960969a88b1cf8413e7c9ec1d5f4955adf91d2d687d8493b70ef457532d505b9cee7a3d2b726a554242b75fb9bec7d4beab74da4bf65260e1d6f7a6b44af4505bf35aaae4cf95b1059ba0f03f1d63c5b7c3ccbacd6bd80577de71f35d0c4976b6e43fe0e1583530e773dfab3ab46c92ce3fa2168673ba52678407a3ef619b5e14155706d43bd329a5e72d36\",\"encoding\":{\"content\":[\"pkcs8\",\"sr25519\"],\"type\":\"xsalsa20-poly1305\",\"version\":\"2\"},\"meta\":{\"name\":\"Yang1\",\"tags\":[],\"whenCreated\":1580628430860}}"
}
```
Success return:
```json
{
    "status": 200
}
```

Failed return example:
```json
{
    "status": 400
}
```

#### Send node get message to get node data (repeatable)
```json
{
    "file_hash": "780f2fe4461952a4fa496127a8bb79bac0957aee2739a3fb84bdc62481db6334",
    "node_hash": "f7197b8762d3a3236f8cefc3d57aaf2811a9225f7ba6490dca6c591ebed4db8c",
    "node_index": 11
}
```

Success return: data (binary)

Failed return example (bad request):
```json
{
    "status": 400
}
```

Failed return example (not found):
```json
{
    "status": 404
}
```
