// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity >=0.8.4 <0.9.0;

import '../libraries/TickMath.sol';

/// @title Test wrapper for TickMath library
/// @notice This contract exposes TickMath library functions for testing
contract TickMathTest {
    /// @notice Exposes TickMath.getSqrtRatioAtTick for testing
    /// @param tick The input tick
    /// @return The sqrt ratio at the given tick
    function getSqrtRatioAtTick(int24 tick) external pure returns (uint160) {
        return TickMath.getSqrtRatioAtTick(tick);
    }

    /// @notice Exposes TickMath.getTickAtSqrtRatio for testing
    /// @param price The sqrt ratio
    /// @return The tick at the given sqrt ratio
    function getTickAtSqrtRatio(uint160 price) external pure returns (int24) {
        return TickMath.getTickAtSqrtRatio(price);
    }

    /// @notice Returns MIN_TICK constant
    function MIN_TICK() external pure returns (int24) {
        return TickMath.MIN_TICK;
    }

    /// @notice Returns MAX_TICK constant
    function MAX_TICK() external pure returns (int24) {
        return TickMath.MAX_TICK;
    }

    /// @notice Returns MIN_SQRT_RATIO constant
    function MIN_SQRT_RATIO() external pure returns (uint160) {
        return TickMath.MIN_SQRT_RATIO;
    }

    /// @notice Returns MAX_SQRT_RATIO constant
    function MAX_SQRT_RATIO() external pure returns (uint160) {
        return TickMath.MAX_SQRT_RATIO;
    }
}
