# README



Avalanche Blackhole Dex에 유동성 공급자로 참여하면서, 자동적으로 리포지셔닝을 해주는 agent입니다.



## 지원 기능

- [x] 토큰 SWAP (WAVAX <=> USDC)

- [ ] 토큰 getPrice

- [ ] STAKE

  - [ ] Mint

  - [ ] Deposit

- [ ] UNSTAKE





## Tx example



### Vote

- txHash: `0x732b789559c8855da5ff26359573dd882cc7d0235e91275b53b32dfe799316d5`
- contractAddr : `0xE30D0C8532721551a51a9FeC7FB233759964d9e3`
- abi : 

- txData: 



### Approve

- txHash: `0x17226fdd0f0df51d1fdd7a47a90de291766f4858a688cdc6c91833b9208bb13f`

- contractAddr : `0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7` (WAVAX)

- abi : `../../blackholedex-contracts/artifacts/@openzeppelin/contracts/token/ERC20/ERC20.sol/ERC20.json`

- txData: ``

- decoded:

  ```json
  {
    "contract": "0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7",
    "method": "approve",
    "signature": "approve(address,uint256)",
    "parameters": [
      {
        "name": "spender",
        "type": "address",
        "value": "0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146"
      },
      {
        "name": "amount",
        "type": "uint256",
        "value": "3750793819555087051"
      }
    ],
    "rawData": "CV6nswAAAAAAAAAAAAAAAD/tAX7A9VF83y6KmkFWxk10JSFGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANA1/GzhPwss="
  }
  ```

  



### Swap

- txHash: `0x1600e68bfd607a5e8452f7533b162eeb4afd4f0435f31639999aa46fbaef79b1`

- contractAddr : `0x04E1dee021Cd12bBa022A72806441B43d8212Fec`

- abi : `../../blackholedex-contracts/artifacts/contracts/RouterV2.sol/RouterV2.json`

- txData: `6ba16543000000000000000000000000000000000000000000000038b4034b62cec2f5a10000000000000000000000000000000000000000000000000000000000000080000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b59223000000000000000000000000000000000000000000000000000000006927fa81000000000000000000000000000000000000000000000000000000000000000100000000000000000000000014e4a5bed2e5e688ee1a5ca3a4914250d1abd573000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000cd94a87696fac69edae3a70fe5725307ae1c43f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b59223`

- decoded

  ```json
  {
    "contract": "0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e",
    "method": "swapExactETHForTokens",
    "signature": "swapExactETHForTokens(uint256,(address,address,address,bool,bool,address)[],address,uint256)",
    "parameters": [
      {
        "name": "amountOutMin",
        "type": "uint256",
        "value": "1045988962367239812513"
      },
      {
        "name": "routes",
        "type": "(address,address,address,bool,bool,address)[]",
        "value": [
          {
            "pair": "0x14e4a5bed2e5e688ee1a5ca3a4914250d1abd573",  
            "from": "0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7",
            "to": "0xcd94a87696fac69edae3a70fe5725307ae1c43f6",
            "stable": false,
            "concentrated": false,
            "receiver": "0xb4dd4fb3d4bced984cce972991fb100488b59223"
          }
        ]
      },
      {
        "name": "to",
        "type": "address",
        "value": "0xb4dd4fb3D4bCED984cce972991fB100488b59223"
      },
      {
        "name": "deadline",
        "type": "uint256",
        "value": "1764227713"
      }
    ],
    "rawData": "a6FlQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOLQDS2LOwvWhAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAC03U+z1LztmEzOlymR+xAEiLWSIwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABpJ/qBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAU5KW+0uXmiO4aXKOkkUJQ0avVcwAAAAAAAAAAAAAAALMfZqo8HnhTY/CHWht04nuF/WbHAAAAAAAAAAAAAAAAzZSodpb6xp7a46cP5XJTB64cQ/YAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAtN1Ps9S87ZhMzpcpkfsQBIi1kiM="
  }
  ```



### Mint NFT (유동성 공급)

- txHash: `0x9e2247a0210448cab301475eef741eba0ee9a9351188a92b8127fce27206b9d0`

- contractAddr : `0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146`

- abi : `../../blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/INonfungiblePositionManager.json`

- txData: ``

- decoded

  ```json
  {
    "contract": "0x3fed017ec0f5517cdf2e8a9a4156c64d74252146",
    "method": "mint",
    "signature": "mint((address,address,address,int24,int24,uint256,uint256,uint256,uint256,address,uint256))",
    "parameters": [
      {
        "name": "params",
        "type": "(address,address,address,int24,int24,uint256,uint256,uint256,uint256,address,uint256)",
        "value": {
          "token0": "0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7",
          "token1": "0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e",
          "deployer": "0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584",
          "tickLower": -249600,
          "tickUpper": -248400,
          "amount0Desired": 3750793819555087051,
          "amount1Desired": 61384000,
          "amount0Min": 3563254128577332698,
          "amount1Min": 58314800,
          "recipient": "0xb4dd4fb3d4bced984cce972991fb100488b59223",
          "deadline": 1763392863027
        }
      }
    ],
    "rawData": "/j875wAAAAAAAAAAAAAAALMfZqo8HnhTY/CHWht04nuF/WbHAAAAAAAAAAAAAAAAuX7574c0xxkE2AAvi2vGbdnEim4AAAAAAAAAAAAAAABdQzqUpKKqj5qjTY0VaS3C6ZYFhP///////////////////////////////////////DEA///////////////////////////////////////8NbAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA0DX8bOE/CywAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADqKVAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMXM4wEJLxdoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA3nQMAAAAAAAAAAAAAAAALTdT7PUvO2YTM6XKZH7EASItZIjAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABmpJnuzM="
  }
  ```

  

### stake
- tx
  - approve : `0xa88b99a8c54187c6d9d078906dd5793f6e9c5354a606d6dc6be719552253ee61`
  - stake(deposit): `0xc4a47d60b5df9f796d10194a03e5f32827cccd2d248a78b626d0a2aafc623401`



### unstake
- tx
  - 0x4e55f91cf25a2bd863027526607eaf62a327d86b0bfb7dedcae31ebcccba179f
- tx data 분석
  ```
  # item 1
  4473eca6                                                            # exitFarming 함수 선택자
  000000000000000000000000cd94a87696fac69edae3a70fe5725307ae1c43f6    # black token
  000000000000000000000000cd94a87696fac69edae3a70fe5725307ae1c43f6    # black token
  00000000000000000000000041100c6d2c6920b10d12cd8d59c8a9aa2ef56fc7    # AlgebraPool
  0000000000000000000000000000000000000000000000000000000000000003    # Nonce
  0000000000000000000000000000000000000000000000000000000000138a9c    # NFT ID

  # item 2
  2f2d783d                                                            # claimReward 함수 선택자
  000000000000000000000000cd94a87696fac69edae3a70fe5725307ae1c43f6    # black token
  000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b59223    # my wallet
  0000000000000000000000000000000000000000000000000000000000000000
  ```


### withdraw
- tx
  - 0x0bee82e46540bd267e86fbc89f3895bd0ce35220c1e1747812801ba854aee6a6
- tx data 분석
  ```
  ac9650d8
  0000000000000000000000000000000000000000000000000000000000000020
  0000000000000000000000000000000000000000000000000000000000000003
  0000000000000000000000000000000000000000000000000000000000000060
  0000000000000000000000000000000000000000000000000000000000000140
  0000000000000000000000000000000000000000000000000000000000000200
  00000000000000000000000000000000000000000000000000000000000000a4

  0c49ccbe                                                            # decreaseLiquidity
  0000000000000000000000000000000000000000000000000000000000138a9c    # nftTokenID
  0000000000000000000000000000000000000000000000000000113c31c1097d    # liquidity
  0000000000000000000000000000000000000000000000000000000000000000    # amount0min
  0000000000000000000000000000000000000000000000000000000000000000    # amount1min
  000000000000000000000000000000000000000000000000000000006947c116    # deadline
  0000000000000000000000000000000000000000000000000000000000000000
  000000000000000000000000000000000000000000000000000000

  84fc6f7865                                                          # collect
  0000000000000000000000000000000000000000000000000000000000138a9c    # nftTokenID
  000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b59223    # myAddr
  00000000000000000000000000000000ffffffffffffffffffffffffffffffff    # Amount0Max
  00000000000000000000000000000000ffffffffffffffffffffffffffffffff    # Amount1Max
  0000000000000000000000000000000000000000000000000000000000000000

  000000000000000000000000000000000000000000000000000000
  2442966c68                                                          # burn
  0000000000000000000000000000000000000000000000000000000000138a9c    # nftTokenID
  00000000000000000000000000000000000000000000000000000000
  ```


## Contracts

### My

| Address                                      | Name         |
| -------------------------------------------- | ------------ |
| `0xb4dd4fb3d4bced984cce972991fb100488b59223` | My Address 1 |




### Blachkhole

| Address                                      | Name                |
| -------------------------------------------- | ------------------- |
| `0x04E1dee021Cd12bBa022A72806441B43d8212Fec` | RouterV             |
| `0xcd94a87696fac69edae3a70fe5725307ae1c43f6` | BLACKHOLE ERC-20    |
| `0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0` | WAVAX/USDC pair     |
| `0x14e4a5bed2e5e688ee1a5ca3a4914250d1abd573` | WAVAX/BLACK pair    |
| `0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584` | mint deployer proxy |



### Tokens

| Address                                      | Name  |
| -------------------------------------------- | ----- |
| `0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7` | WAVAX |
| `0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E` | USDC  |




## logs

### mint
2025/12/15 19:18:07 
ApproveWAVAX: 0x8b001614364317382c0f1f611f90b89be9237fb6f135ad96f01689577d82b832 (gas: 45341693835775 wei)
Mint: 0x1c4fe48227c85da5a970a87b08540b56f0cb5e0779f2adcbba0903e0c25c7c1b (gas: 627485601390516 wei)
packed : e3f3be7000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000b97ef9ef8734c71904d8002f8b6bc66dd9c48a6e0000000000000000000000005d433a94a4a2aa8f9aa34d8d15692dc2e9960584fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc29f8fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc31c80000000000000000000000000000000000000000000000000de0b6b3a763daad0000000000000000000000000000000000000000000000000000000000a69ad10000000000000000000000000000000000000000000000000d2f13f7789edc8a00000000000000000000000000000000000000000000000000000000009e4646000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b5922300000000000000000000000000000000000000000000000000000000693fe515
