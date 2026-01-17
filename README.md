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
            "concentrated": false, // 이건 true로 세팅해서 사용하기
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
  - stable : determines the mathematical formula (invariant) used for the swap in Basic Pools (V2-style).
    - false (Volatile): Uses the standard Constant Product Formula ($x \times y = k$). This is designed for assets that fluctuate in price relative to each other 
    - true (Stable): Uses a StableSwap Invariant (similar to Curve, e.g., $x^3y + y^3x = k$). This is optimized for correlated assets that should stay at a 1:1 price ratio, providing much lower slippage.
  - concentrated : whether to use the Concentrated Liquidity engine (V3-style) instead of a standard V2-style pool.
    - false: The router looks for a "Basic Pool" where liquidity is distributed infinitely across the entire price curve (from 0 to infinity).
    - true: The router looks for a Concentrated Liquidity Pool. In these pools, liquidity is provided within specific price ranges (ticks)

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


---

84455@84455 blackhole_dex % go run cmd/main.go 
2025/12/26 08:44:02 Strategy report channel full, dropping message: strategy_start
{"timestamp":"2025-12-26T08:44:02.65556+09:00","event_type":"strategy_start","message":"RunStrategy1 starting - automated liquidity repositioning","phase":0}
2025/12/26 08:44:03 CalculateRebalanceAmounts: WAVAX 5023780141631114555, USDC 75496630, price : 3306379361727413336
2025/12/26 08:44:03 Result of CalculateRebalanceAmounts: direction 0,swapAmount : 2511890070812443883
{"timestamp":"2025-12-26T08:44:03.508349+09:00","event_type":"swap_complete","message":"Rebalancing: swapping token 0 amount 2511890070812443883","phase":0}
packed : 095ea7b300000000000000000000000004e1dee021cd12bba022a72806441b43d8212fec00000000000000000000000000000000000000000000000022dc06b5f99b44eb
packed : 204b5c0a00000000000000000000000000000000000000000000000022dc06b5f99b44eb0000000000000000000000000000000000000000000000000000000001b97a0b00000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b5922300000000000000000000000000000000000000000000000000000000694dd0f30000000000000000000000000000000000000000000000000000000000000001000000000000000000000000a02ec3ba8d17887567672b2cdcaf525534636ea0000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000b97ef9ef8734c71904d8002f8b6bc66dd9c48a6e00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b59223
{"timestamp":"2025-12-26T08:44:18.86769+09:00","event_type":"gas_cost","message":"Swap transaction completed","phase":0,"gas_cost":759445521202380,"cumulative_gas":759445521202380}
{"timestamp":"2025-12-26T08:44:19.459901+09:00","event_type":"position_created","message":"Minting position with RangeWidth 10","phase":0}
2025/12/26 08:44:19 Capital Utilization: WAVAX 99%, USDC 20%
2025/12/26 08:44:19 ⚠️  Capital Efficiency Warning: 79% of USDC (84192684 smallest unit) will not be staked. Consider adjusting amounts or range width.
packed : fe3f3be7000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000b97ef9ef8734c71904d8002f8b6bc66dd9c48a6e0000000000000000000000005d433a94a4a2aa8f9aa34d8d15692dc2e9960584fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc26d8fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc2ea800000000000000000000000000000000000000000000000022dc06b5f9fa2a3c00000000000000000000000000000000000000000000000000000000014bc9ca000000000000000000000000000000000000000000000000211dd32ce0ada81f00000000000000000000000000000000000000000000000000000000013b32e6000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b5922300000000000000000000000000000000000000000000000000000000694dd104
event.ID.Hex(): 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31 | log.Topics[0].Hex(): 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
event.ID.Hex(): 0x40d0efd1a53d60ecbf40971b9daf7dc90178c3aadc7aab1765632738fa8b8f01 | log.Topics[0].Hex(): 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
event.ID.Hex(): 0x26f6a048ee9138f2c0ce266f322cb99228e8d619ae2bff30c67f8dcf9d2377b4 | log.Topics[0].Hex(): 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
event.ID.Hex(): 0x4f27462fbdc9bce16bb573a06acba6b27394e151da96ce8098d8e29a6dc8d64b | log.Topics[0].Hex(): 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
event.ID.Hex(): 0x8a82de7fe9b33e0e6bca0e26f5bd14a74f1164ffe236d50e0a36c3ea70f2b814 | log.Topics[0].Hex(): 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
event.ID.Hex(): 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef | log.Topics[0].Hex(): 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
panic: runtime error: index out of range [2] with length 2

goroutine 23 [running]:
blackholego/pkg/contractclient.(*ContractClient).ParseReceipt(0x140000ae030, 0x1400014e488)
        /Users/84455/workspace/blackhole_dex/pkg/contractclient/contractclient.go:311 +0x84c
blackholego.MintNftTokenId({0x102907f38, 0x140000ae030}, 0x1400014e488)
        /Users/84455/workspace/blackhole_dex/blackhole.go:1251 +0x60
blackholego.(*Blackhole).Mint(0x1400044e000, 0x14000311200, 0x14000201360, 0xa, 0x5)
        /Users/84455/workspace/blackhole_dex/blackhole.go:579 +0x2064
blackholego.(*Blackhole).initialPositionEntry(0x1400044e000, 0x140000ce000, 0x1400013c120, 0x1400044a070)
        /Users/84455/workspace/blackhole_dex/blackhole.go:1542 +0xee0
blackholego.(*Blackhole).RunStrategy1(0x1400044e000, {0x102904fd0, 0x102bc7d20}, 0x1400044a070, 0x140000ce000)
        /Users/84455/workspace/blackhole_dex/blackhole.go:2082 +0x268
created by main.main in goroutine 1
        /Users/84455/workspace/blackhole_dex/cmd/main.go:60 +0x3c0


## 이슈 기록

### RPC State Lag (Node Desync)
- 현상
  - 사전 트랜잭션에 대한 receipt까지 받은 후 후속 요청을 보냈지만 실패
  - 잠시 기다렸다가 시도 시 성공
- 원인
  - The Load Balancer "Desync"
      1. Most public RPC endpoints (like api.avax.network) use a load balancer that sits in front of dozens of different nodes.
      2. Transaction 1: Hits Node A. Node A processes it, includes it in a block, and gives you a success receipt.
      3. Transaction 2: You send it immediately. The load balancer might route this request to Node B.
    => The Problem: Node B might be a few milliseconds behind Node A. It hasn't "seen" the block containing your first transaction yet. If Transaction 2 depends on the state changed in Transaction 1 (like a balance update or a contract flag), Node B will reject it as invalid.
    :bulb: EstimateGas 단계에서 에러가 발생하는 것이라 nonce와 무관하게 에러 발생.


### execution reverted: STF
- 개요
  - a specific short-code used in Uniswap V3-style contracts
  - Safe Transfer Failed의 약어
- common cause
  - Insufficient Balance
  - Incomplete Approval
  - RPC State Lag