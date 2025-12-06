const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("TickMath Library", function () {
  let tickMathTest;

  beforeEach(async function () {
    const TickMathTest = await ethers.getContractFactory("TickMathTest");
    tickMathTest = await TickMathTest.deploy();
    await tickMathTest.deployed();
  });

  describe("Constants", function () {
    it("should return correct MIN_TICK", async function () {
      expect(await tickMathTest.MIN_TICK()).to.equal(-887272);
    });

    it("should return correct MAX_TICK", async function () {
      expect(await tickMathTest.MAX_TICK()).to.equal(887272);
    });

    it("should return correct MIN_SQRT_RATIO", async function () {
      expect(await tickMathTest.MIN_SQRT_RATIO()).to.equal("4295128739");
    });

    it("should return correct MAX_SQRT_RATIO", async function () {
      expect(await tickMathTest.MAX_SQRT_RATIO()).to.equal(
        "1461446703485210103287273052203988822378723970342"
      );
    });
  });

  describe("getSqrtRatioAtTick", function () {
    it("should return MIN_SQRT_RATIO for MIN_TICK", async function () {
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(-887272);
      expect(sqrtRatio).to.equal(await tickMathTest.MIN_SQRT_RATIO());
    });

    it("should return MAX_SQRT_RATIO for MAX_TICK", async function () {
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(887272);
      expect(sqrtRatio).to.equal(await tickMathTest.MAX_SQRT_RATIO());
    });

    it("should calculate correct ratio for tick 0", async function () {
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(0);
      // At tick 0, the sqrt ratio should be approximately 2^96 (equal price)
      const expected = ethers.BigNumber.from(2).pow(96);
      // Allow for small rounding difference
      expect(sqrtRatio).to.be.closeTo(expected, ethers.BigNumber.from(10).pow(10));
    });

    it("should revert for tick > MAX_TICK", async function () {
      await expect(
        tickMathTest.getSqrtRatioAtTick(887273)
      ).to.be.reverted;
    });

    it("should revert for tick < MIN_TICK", async function () {
      await expect(
        tickMathTest.getSqrtRatioAtTick(-887273)
      ).to.be.reverted;
    });

    it("should handle negative ticks", async function () {
      const tick = -1000;
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(tick);
      expect(sqrtRatio).to.be.gt(0);
    });

    it("should handle positive ticks", async function () {
      const tick = 1000;
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(tick);
      expect(sqrtRatio).to.be.gt(0);
    });

    it("should calculate correct ratio for tick x", async function () {
      const x = -249428;
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(x);
      // Expected sqrt ratio for tick -249428
      const expected = ethers.BigNumber.from("304011615425126403287043");
      // Allow for small rounding difference
      expect(sqrtRatio).to.be.closeTo(expected, ethers.BigNumber.from(10).pow(10));
    });
  });

  describe("getTickAtSqrtRatio", function () {
    it("should return MIN_TICK for MIN_SQRT_RATIO", async function () {
      const tick = await tickMathTest.getTickAtSqrtRatio("4295128739");
      expect(tick).to.equal(-887272);
    });

    it("should revert for price < MIN_SQRT_RATIO", async function () {
      await expect(
        tickMathTest.getTickAtSqrtRatio("4295128738")
      ).to.be.reverted;
    });

    it("should revert for price >= MAX_SQRT_RATIO", async function () {
      await expect(
        tickMathTest.getTickAtSqrtRatio("1461446703485210103287273052203988822378723970342")
      ).to.be.reverted;
    });

    it("should be inverse of getSqrtRatioAtTick", async function () {
      const tick = -249428;
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(tick);
      const tickBack = await tickMathTest.getTickAtSqrtRatio(sqrtRatio);
      expect(tickBack).to.equal(tick);
    });

    it("should handle various sqrt ratios correctly", async function () {
      const testTicks = [-100000, -1000, -1, 0, 1, 1000, 100000];

      for (const tick of testTicks) {
        const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(tick);
        const tickBack = await tickMathTest.getTickAtSqrtRatio(sqrtRatio);
        expect(tickBack).to.equal(tick);
      }
    });

    it("should be equal to expected tick", async function () {
      const tick = -249428;
      const sqrtRatio = ethers.BigNumber.from("304014154377809408260091");
      const tickBack = await tickMathTest.getTickAtSqrtRatio(sqrtRatio);
      expect(tickBack).to.equal(tick);
    });
  });

  describe("Various tick values", function () {
    const testCases = [
      { tick: -100000, description: "large negative tick" },
      { tick: -1000, description: "medium negative tick" },
      { tick: -1, description: "small negative tick" },
      { tick: 0, description: "zero tick" },
      { tick: 1, description: "small positive tick" },
      { tick: 1000, description: "medium positive tick" },
      { tick: 100000, description: "large positive tick" },
    ];

    testCases.forEach(({ tick, description }) => {
      it(`should handle ${description} (${tick})`, async function () {
        const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(tick);
        expect(sqrtRatio).to.be.gt(0);

        const tickBack = await tickMathTest.getTickAtSqrtRatio(sqrtRatio);
        expect(tickBack).to.equal(tick);
      });
    });
  });

  describe("Boundary conditions", function () {
    it("should handle tick just below MAX_TICK", async function () {
      const tick = 887271;
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(tick);
      expect(sqrtRatio).to.be.gt(0);
      expect(sqrtRatio).to.be.lt(await tickMathTest.MAX_SQRT_RATIO());
    });

    it("should handle tick just above MIN_TICK", async function () {
      const tick = -887271;
      const sqrtRatio = await tickMathTest.getSqrtRatioAtTick(tick);
      expect(sqrtRatio).to.be.gt(await tickMathTest.MIN_SQRT_RATIO());
    });
  });

  describe("Monotonicity", function () {
    it("should produce increasing sqrt ratios for increasing ticks", async function () {
      const tick1 = -1000;
      const tick2 = 0;
      const tick3 = 1000;

      const sqrtRatio1 = await tickMathTest.getSqrtRatioAtTick(tick1);
      const sqrtRatio2 = await tickMathTest.getSqrtRatioAtTick(tick2);
      const sqrtRatio3 = await tickMathTest.getSqrtRatioAtTick(tick3);

      expect(sqrtRatio1).to.be.lt(sqrtRatio2);
      expect(sqrtRatio2).to.be.lt(sqrtRatio3);
    });
  });
});
