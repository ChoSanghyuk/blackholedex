# README

Avalanche Blackhole Dex에 유동성 공급자로 참여하면서, 자동적으로 리포지셔닝을 해주는 agent입니다.

## RunStrategy1 전략 동작 방식

### 개요
`RunStrategy1`은 Blackhole DEX에서 자동화된 유동성 리포지셔닝 전략을 실행하는 메인 함수입니다. 

### 핵심 User Story
1. **US1**: 자동 리밸런싱을 포함한 초기 포지션 진입
2. **US2**: 지속적인 가격 모니터링
3. **US3**: 범위 이탈 시 자동 포지션 리밸런싱
4. **US4**: 재진입 전 가격 안정성 감지

### 전략 단계 (StrategyPhase)

#### 1. Initializing (초기화)
- 새로운 유동성 포지션 생성 단계
- `initialPositionEntry()` 함수 실행:
  - WAVAX/USDC 잔액 확인
  - 풀 가격 조회
  - 필요시 리밸런싱 스왑 실행 (0.1 AVAX 또는 1 USDC 이상일 때만)
  - NFT 민팅 (유동성 공급)
  - 스테이킹 (인센티브 프로그램 참여)
- 성공 시 → **ActiveMonitoring** 단계로 전환

#### 2. ActiveMonitoring (활성 모니터링)
- 풀 가격을 주기적으로 모니터링 (`monitoringLoop()`)
- 현재 틱이 포지션 범위 내에 있는지 확인:
  - `tickLower ≤ currentTick ≤ tickUpper`
- 범위 이탈 감지 시 → **RebalancingRequired** 단계로 전환

#### 3. RebalancingRequired (리밸런싱 필요)
- `executeRebalancing()` 함수 실행:
  1. **Unstake**: 스테이킹된 NFT 회수 및 보상 수령
  2. **Withdraw**: 유동성 제거 및 NFT 소각
     - `decreaseLiquidity()`: 모든 유동성 제거
     - `collect()`: 토큰 회수
     - `burn()`: NFT 소각
- 성공 시 → **WaitingForStability** 단계로 전환

#### 4. WaitingForStability (안정성 대기)
- `stabilityLoop()` 함수로 가격 안정성 체크
- **StabilityWindow** 사용:
  - 이전 가격과 현재 가격 비교
  - 변동폭이 임계값(`StabilityThreshold`) 이하인지 확인
  - 필요한 횟수(`StabilityIntervals`)만큼 안정 상태 유지 필요
- 가격 변동성이 큰 경우 안정성 카운터 초기화
- 안정화 완료 시 → **Initializing** 단계로 전환하여 재진입

#### 5. Halted (중단)
- 치명적 오류 발생 시 진입하는 안전 상태
- **CircuitBreaker**가 오류 패턴 감지:
  - 짧은 시간 내 반복적 오류
  - 치명적 오류 발생 (RPC 연결 실패, 심각한 컨트랙트 오류 등)
- 전략 종료 및 최종 리포트 생성

### 주요 설정 (StrategyConfig)

| 파라미터 | 설명 |
|---------|------|
| `MonitoringInterval` | 모니터링 주기 (예: 30초) |
| `RangeWidth` | 포지션 틱 범위 너비 |
| `SlippagePct` | 슬리피지 허용 비율 (예: 5%) |
| `StabilityThreshold` | 가격 안정성 임계값 |
| `StabilityIntervals` | 필요한 안정 구간 횟수 |
| `CircuitBreakerWindow` | 오류 감지 시간 창 |
| `CircuitBreakerThreshold` | 중단 트리거 오류 횟수 |

### 자동 스냅샷 기록
- 전략 시작 시 초기 자산 스냅샷 기록
- 2시간마다 자동 스냅샷 (WAVAX, USDC, BLACK, AVAX 잔액)
- 각 단계 완료 시 스냅샷 기록 (Initializing, RebalancingRequired 완료 시)
- 포지션 내 유동성 가치도 잔액에 포함하여 총 자산 계산

### 상태 추적 (StrategyState)
- `NFTTokenID`: 현재 포지션 NFT ID
- `TickLower/TickUpper`: 포지션 범위
- `CumulativeGas`: 누적 가스 비용
- `CumulativeRewards`: 누적 보상
- `TotalSwapFees`: 총 스왑 수수료
- `LastPrice`: 마지막 관찰 가격
- `PositionCreatedAt`: 포지션 생성 시각

### 리포팅 시스템
전략 실행 중 다음 이벤트 발생 시 리포트 생성:
- `strategy_start`: 전략 시작
- `position_created`: 포지션 생성 완료
- `position_loaded`: 기존 포지션 로드
- `monitoring`: 가격 모니터링 (로그로만 기록)
- `out_of_range`: 범위 이탈 감지
- `rebalance_start`: 리밸런싱 시작
- `stability_check`: 안정성 체크 진행 상황
- `error`: 오류 발생
- `shutdown`: 전략 종료



## 지원 기능



 ###  주요 트랜잭션 함수

- [x] Swap :  토큰 간 스왑 실행 (WAVAX ↔ USDC 등)
- [x]  Mint :  WAVAX-USDC 풀에 유동성 공급 (NFT 생성)
- [x] Stake :  유동성 포지션 NFT를 스테이킹
- [x] Unstake :  스테이킹된 NFT 회수
- [x] Withdraw : 포지션에서 모든 유동성 제거 및 NFT 소각

### 조회 함수

- [x] GetAMMState : AMM 풀의 현재 상태 조회
- [x] GetUserPositions : 사용자가 소유한 모든 NFT 포지션 ID 조회
- [x] GetPositionDetails : 특정 NFT 포지션의 상세 정보 조회
- [x] TokenOfOwnerByIndex : 인덱스로 사용자의 NFT 토큰 ID 조회





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





## 수익률 비교 시나리오

초기 세팅 값을 가정했을 때, 가격의 하락과 상승에 따라 공급한 유동성 pool에서 가격이 이탈한 경우의 수익률 비교 시나리오입니다.

시나리오에 대한 시뮬레이션은 `pkg/util.simulation_test.go`의 `TestPriceMovementSimulation` 테스트 함수에서 진행 가능합니다.

### 초기 세팅 

#### TICK 설정

| tick spacing | tick width |   tick range   |
| :----------: | :--------: | :------------: |
|     200      |     6      | 1200 (6 * 200) |

#### 현재가

| 현재 tick | 하한 tick | 상단 tick |
| :-------: | :-------: | :-------: |
|  -251400  |  -252000  |  -250800  |

| 현재가 (USDC/AVAX) | 하한가(USDC/AVAX) | 상단가(USDC/AVAX) |
| :----------------: | :---------------: | :---------------: |
|      12.0888       |      11.3848      |      12.8363      |

#### 초기 자본

| 총 자산 가치 | AVAX 수량 | USDC 수량 |
| :----------: | :-------: | :-------: |
| 1000.00 USD  | 41.360655 |  500.00   |



### 시나리오 1. 하한가까지 가격 하락

초기 자본이 DEX에 공급되어 있을 때, `tick`(가격)이 `-251400`(12.0888)에서 `-252000`(11.3848)로 이동했을 때의 자산 변화

#### 시나리오1 자본 변동 추이

| 총 자산 가치 | AVAX 수량 | USDC 수량 |
| :----------: | :-------: | :-------: |
|  956.11 USD  | 83.980866 |     0     |

손실액은 **43.89 USD**으로 전체액에서 **4.39%** 손실을 입게 된다.

#### DEX 미참여 시의 자본 변동 추이

| 총 자산 가치 | AVAX 수량 | USDC 수량 |
| :----------: | :-------: | :-------: |
|  970.88 USD  | 41.360655 |  500.00   |

유동성 공급을 하지 않고 초기 자본을 그대로 들고 있을 때와 비교해서는 **14.78 USD**만큼 더 손실을 입게 된다. 



### 시나리오 2. 상한가까지 가격 상승

초기 자본이 DEX에 공급되어 있을 때, `tick`(가격)이 `-251400`(12.0888)에서 `-250800`(12.8363)로 이동했을 때의 자산 변화

#### 시나리오2 자본 변동 추이

| 총 자산 가치 | AVAX 수량 | USDC 수량 |
| :----------: | :-------: | :-------: |
| 1015.23 USD  |     0     |  1015.23  |

수익액은 **15.23 USD**으로 전체액에서 **1.52%** 수익을 얻게 된다.

#### DEX 미참여 시의 자본 변동 추이

| 총 자산 가치 | AVAX 수량 | USDC 수량 |
| :----------: | :-------: | :-------: |
| 1030.92 USD  | 41.360655 |  500.00   |

유동성 공급을 하지 않고 초기 자본을 그대로 들고 있을 때와 비교해서는 **15.69 USD**만큼 덜 수익을 얻게된다.



### 시나리오 3. 가격 변동에 따른 리포지션닝 후 가격 회복

초기 자본이 DEX에 공급되어 있을 때, 가격 하락으로 가격이 기존 포지션을 이탈하여 리포지셔닝을 수행했지만, 다시 원래 가격으로 회복했을 때의 시나리오

다음과 같은 과정을 걸쳤을 때의 시나리오

1. `tick`(가격)이 `-251400`(12.0888)에서 `-250800`(12.8363)로 이동
2. DEX에서 유동성 포지션 회수 및 AVAX와 USDC의 비중을 50:50으로 맞춘 후 다시 재공급
3. `tick`(가격)이  `-250800`(12.8363)에서 `-251400`(12.0888)로 회복

#### 시나리오3 자본 변동 추이

| 총 자산 가치 | AVAX 수량 | USDC 수량 |
| :----------: | :-------: | :-------: |
|  970.67 USD  |     0     |  970.67   |

손실액은 **29.33 USD**으로 전체액에서 **2.93%** 손실을 입게 된다. 



### 결론

width를 6으로 선정했을 때에는, DEX에 미참여했을 때와 비교했을 때 대략 **1.5**% 손실 위험이 따르는 것을 알 수 있다. 또한 일반적인 DEX의 참여와 달리 해당 프로젝트의 사용으로 적극적인 리포지션을 채택했을 때에는 DEX의 일반적인 비영구적인 손실이 아닌 영구적인 손실을 입게 된다. 

따라서, 가격이 변동할 때에는 적극적인 리포지션을 통해 손실액보다 큰 incentive를 받고 있는지 지속적인 모니터링이 필요하며, 이를 기반으로 공급할 width를 조절해야 할 것이다.



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



## Tx example



### Vote

- txHash: `0x732b789559c8855da5ff26359573dd882cc7d0235e91275b53b32dfe799316d5`
- contractAddr : `0xE30D0C8532721551a51a9FeC7FB233759964d9e3`



### Approve

- txHash: `0x17226fdd0f0df51d1fdd7a47a90de291766f4858a688cdc6c91833b9208bb13f`

- contractAddr : `0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7` (WAVAX)

- abi : `../../blackholedex-contracts/artifacts/@openzeppelin/contracts/token/ERC20/ERC20.sol/ERC20.json`

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

  - **stable** : determines the mathematical formula (invariant) used for the swap in Basic Pools (V2-style).
    - false (Volatile): Uses the standard Constant Product Formula ($x \times y = k$). This is designed for assets that fluctuate in price relative to each other 
    - true (Stable): Uses a StableSwap Invariant (similar to Curve, e.g., $x^3y + y^3x = k$). This is optimized for correlated assets that should stay at a 1:1 price ratio, providing much lower slippage.
  - **concentrated** : whether to use the Concentrated Liquidity engine (V3-style) instead of a standard V2-style pool.
    - false: The router looks for a "Basic Pool" where liquidity is distributed infinitely across the entire price curve (from 0 to infinity).
    - true: The router looks for a Concentrated Liquidity Pool. In these pools, liquidity is provided within specific price ranges (ticks)

### Mint NFT (유동성 공급)

- txHash: `0x9e2247a0210448cab301475eef741eba0ee9a9351188a92b8127fce27206b9d0`

- contractAddr : `0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146`

- abi : `../../blackholedex-contracts/artifacts/@cryptoalgebra/integral-periphery/contracts/interfaces/INonfungiblePositionManager.sol/INonfungiblePositionManager.json`

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

