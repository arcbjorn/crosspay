// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";

contract TrancheVault is ERC20, Ownable, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    enum TrancheType {
        Junior,
        Mezzanine, 
        Senior
    }

    struct TrancheInfo {
        uint256 totalDeposits;
        uint256 currentBalance;
        uint256 yieldRate;
        uint256 riskMultiplier;
        uint256 lastYieldUpdate;
        bool isActive;
    }

    struct UserPosition {
        uint256 juniorDeposit;
        uint256 mezzanineDeposit;
        uint256 seniorDeposit;
        uint256 lastDepositTime;
        uint256 accruedYield;
        uint256 lastYieldCalculation;
    }

    struct SlashingEvent {
        uint256 amount;
        uint256 timestamp;
        string reason;
        address validator;
        uint256 juniorSlashed;
        uint256 mezzanineSlashed;
        uint256 seniorSlashed;
    }

    IERC20 public immutable depositToken;
    
    mapping(TrancheType => TrancheInfo) public tranches;
    mapping(address => UserPosition) public userPositions;
    mapping(uint256 => SlashingEvent) public slashingEvents;
    
    uint256 public totalVaultAssets;
    uint256 public insuranceFund;
    uint256 public performanceFeeRate = 1000; // 10%
    uint256 public slashingEventCounter;
    
    uint256 public constant JUNIOR_YIELD_BASE = 1200; // 12%
    uint256 public constant MEZZANINE_YIELD_BASE = 800; // 8%
    uint256 public constant SENIOR_YIELD_BASE = 500; // 5%
    
    uint256 public constant JUNIOR_RISK_MULTIPLIER = 300; // 3x
    uint256 public constant MEZZANINE_RISK_MULTIPLIER = 200; // 2x
    uint256 public constant SENIOR_RISK_MULTIPLIER = 100; // 1x

    uint256 public constant MIN_DEPOSIT = 1e18; // 1 token
    uint256 public constant WITHDRAWAL_DELAY = 7 days;
    
    mapping(address => uint256) public withdrawalRequests;
    mapping(address => uint256) public withdrawalRequestTime;

    event Deposited(
        address indexed user,
        TrancheType indexed tranche,
        uint256 amount,
        uint256 shares
    );

    event WithdrawalRequested(
        address indexed user,
        TrancheType indexed tranche,
        uint256 amount
    );

    event Withdrawn(
        address indexed user,
        TrancheType indexed tranche,
        uint256 amount,
        uint256 shares
    );

    event YieldDistributed(
        TrancheType indexed tranche,
        uint256 totalYield,
        uint256 perTokenYield
    );

    event Slashed(
        uint256 indexed eventId,
        uint256 totalAmount,
        uint256 juniorLoss,
        uint256 mezzanineLoss,
        uint256 seniorLoss,
        address validator,
        string reason
    );

    event TrancheRebalanced(
        TrancheType indexed tranche,
        uint256 oldBalance,
        uint256 newBalance
    );

    event PerformanceFeeCollected(
        uint256 amount,
        address indexed collector
    );

    error InsufficientDeposit();
    error TrancheNotActive();
    error InsufficientBalance();
    error WithdrawalNotRequested();
    error WithdrawalDelayNotMet();
    error InvalidTrancheType();
    error SlashingFailed();
    error RebalancingFailed();

    constructor(
        address _depositToken,
        string memory _name,
        string memory _symbol
    ) ERC20(_name, _symbol) Ownable(msg.sender) {
        depositToken = IERC20(_depositToken);
        
        _initializeTranches();
    }

    function _initializeTranches() internal {
        tranches[TrancheType.Junior] = TrancheInfo({
            totalDeposits: 0,
            currentBalance: 0,
            yieldRate: JUNIOR_YIELD_BASE,
            riskMultiplier: JUNIOR_RISK_MULTIPLIER,
            lastYieldUpdate: block.timestamp,
            isActive: true
        });

        tranches[TrancheType.Mezzanine] = TrancheInfo({
            totalDeposits: 0,
            currentBalance: 0,
            yieldRate: MEZZANINE_YIELD_BASE,
            riskMultiplier: MEZZANINE_RISK_MULTIPLIER,
            lastYieldUpdate: block.timestamp,
            isActive: true
        });

        tranches[TrancheType.Senior] = TrancheInfo({
            totalDeposits: 0,
            currentBalance: 0,
            yieldRate: SENIOR_YIELD_BASE,
            riskMultiplier: SENIOR_RISK_MULTIPLIER,
            lastYieldUpdate: block.timestamp,
            isActive: true
        });
    }

    function deposit(
        TrancheType tranche,
        uint256 amount
    ) external nonReentrant whenNotPaused {
        if (amount < MIN_DEPOSIT) {
            revert InsufficientDeposit();
        }
        if (!tranches[tranche].isActive) {
            revert TrancheNotActive();
        }

        _updateYield(tranche);

        depositToken.safeTransferFrom(msg.sender, address(this), amount);

        UserPosition storage position = userPositions[msg.sender];
        
        if (tranche == TrancheType.Junior) {
            position.juniorDeposit += amount;
        } else if (tranche == TrancheType.Mezzanine) {
            position.mezzanineDeposit += amount;
        } else {
            position.seniorDeposit += amount;
        }

        position.lastDepositTime = block.timestamp;
        position.lastYieldCalculation = block.timestamp;

        tranches[tranche].totalDeposits += amount;
        tranches[tranche].currentBalance += amount;
        totalVaultAssets += amount;

        uint256 shares = _calculateShares(amount, tranche);
        _mint(msg.sender, shares);

        emit Deposited(msg.sender, tranche, amount, shares);
    }

    function requestWithdrawal(
        TrancheType tranche,
        uint256 amount
    ) external {
        UserPosition storage position = userPositions[msg.sender];
        uint256 userDeposit = _getUserTrancheDeposit(position, tranche);
        
        if (userDeposit < amount) {
            revert InsufficientBalance();
        }

        withdrawalRequests[msg.sender] = amount;
        withdrawalRequestTime[msg.sender] = block.timestamp;

        emit WithdrawalRequested(msg.sender, tranche, amount);
    }

    function withdraw(TrancheType tranche) external nonReentrant whenNotPaused {
        if (withdrawalRequests[msg.sender] == 0) {
            revert WithdrawalNotRequested();
        }
        if (block.timestamp < withdrawalRequestTime[msg.sender] + WITHDRAWAL_DELAY) {
            revert WithdrawalDelayNotMet();
        }

        _updateYield(tranche);

        uint256 amount = withdrawalRequests[msg.sender];
        UserPosition storage position = userPositions[msg.sender];

        uint256 totalWithdrawable = amount + position.accruedYield;
        
        if (tranche == TrancheType.Junior) {
            position.juniorDeposit -= amount;
        } else if (tranche == TrancheType.Mezzanine) {
            position.mezzanineDeposit -= amount;
        } else {
            position.seniorDeposit -= amount;
        }

        position.accruedYield = 0;
        position.lastYieldCalculation = block.timestamp;

        tranches[tranche].totalDeposits -= amount;
        tranches[tranche].currentBalance -= amount;
        totalVaultAssets -= amount;

        uint256 shares = _calculateShares(amount, tranche);
        _burn(msg.sender, shares);

        withdrawalRequests[msg.sender] = 0;
        withdrawalRequestTime[msg.sender] = 0;

        depositToken.safeTransfer(msg.sender, totalWithdrawable);

        emit Withdrawn(msg.sender, tranche, totalWithdrawable, shares);
    }

    function executeSlashing(
        uint256 slashAmount,
        address validator,
        string calldata reason
    ) external onlyOwner {
        if (slashAmount > totalVaultAssets) {
            revert SlashingFailed();
        }

        uint256 remainingSlash = slashAmount;
        uint256 juniorSlashed = 0;
        uint256 mezzanineSlashed = 0;
        uint256 seniorSlashed = 0;

        if (remainingSlash > 0 && tranches[TrancheType.Junior].currentBalance > 0) {
            uint256 juniorAvailable = tranches[TrancheType.Junior].currentBalance;
            juniorSlashed = remainingSlash > juniorAvailable ? juniorAvailable : remainingSlash;
            tranches[TrancheType.Junior].currentBalance -= juniorSlashed;
            remainingSlash -= juniorSlashed;
        }

        if (remainingSlash > 0 && tranches[TrancheType.Mezzanine].currentBalance > 0) {
            uint256 mezzanineAvailable = tranches[TrancheType.Mezzanine].currentBalance;
            mezzanineSlashed = remainingSlash > mezzanineAvailable ? mezzanineAvailable : remainingSlash;
            tranches[TrancheType.Mezzanine].currentBalance -= mezzanineSlashed;
            remainingSlash -= mezzanineSlashed;
        }

        if (remainingSlash > 0 && tranches[TrancheType.Senior].currentBalance > 0) {
            uint256 seniorAvailable = tranches[TrancheType.Senior].currentBalance;
            seniorSlashed = remainingSlash > seniorAvailable ? seniorAvailable : remainingSlash;
            tranches[TrancheType.Senior].currentBalance -= seniorSlashed;
            remainingSlash -= seniorSlashed;
        }

        slashingEventCounter++;
        slashingEvents[slashingEventCounter] = SlashingEvent({
            amount: slashAmount,
            timestamp: block.timestamp,
            reason: reason,
            validator: validator,
            juniorSlashed: juniorSlashed,
            mezzanineSlashed: mezzanineSlashed,
            seniorSlashed: seniorSlashed
        });

        totalVaultAssets -= (slashAmount - remainingSlash);

        emit Slashed(
            slashingEventCounter,
            slashAmount,
            juniorSlashed,
            mezzanineSlashed,
            seniorSlashed,
            validator,
            reason
        );
    }

    function distributeYield() external onlyOwner {
        _updateYield(TrancheType.Junior);
        _updateYield(TrancheType.Mezzanine);
        _updateYield(TrancheType.Senior);
    }

    function _updateYield(TrancheType tranche) internal {
        TrancheInfo storage info = tranches[tranche];
        
        if (info.totalDeposits == 0) {
            info.lastYieldUpdate = block.timestamp;
            return;
        }

        uint256 timeElapsed = block.timestamp - info.lastYieldUpdate;
        if (timeElapsed == 0) return;

        uint256 yieldAmount = (info.currentBalance * info.yieldRate * timeElapsed) / (365 days * 10000);
        
        uint256 performanceFee = (yieldAmount * performanceFeeRate) / 10000;
        uint256 netYield = yieldAmount - performanceFee;

        info.currentBalance += netYield;
        totalVaultAssets += netYield;
        insuranceFund += performanceFee;
        info.lastYieldUpdate = block.timestamp;

        emit YieldDistributed(tranche, netYield, netYield * 1e18 / info.totalDeposits);
    }

    function rebalanceTranches() external onlyOwner {
        uint256 totalBalance = totalVaultAssets;
        
        uint256 targetJunior = totalBalance * 20 / 100;   // 20%
        uint256 targetMezzanine = totalBalance * 30 / 100; // 30%  
        uint256 targetSenior = totalBalance * 50 / 100;   // 50%

        _rebalanceTranche(TrancheType.Junior, targetJunior);
        _rebalanceTranche(TrancheType.Mezzanine, targetMezzanine);
        _rebalanceTranche(TrancheType.Senior, targetSenior);
    }

    function _rebalanceTranche(TrancheType tranche, uint256 targetBalance) internal {
        TrancheInfo storage info = tranches[tranche];
        uint256 oldBalance = info.currentBalance;

        if (oldBalance != targetBalance) {
            info.currentBalance = targetBalance;
            emit TrancheRebalanced(tranche, oldBalance, targetBalance);
        }
    }

    function _calculateShares(uint256 amount, TrancheType tranche) internal view returns (uint256) {
        TrancheInfo storage info = tranches[tranche];
        
        if (info.totalDeposits == 0) {
            return amount;
        }

        return (amount * totalSupply()) / info.totalDeposits;
    }

    function _getUserTrancheDeposit(
        UserPosition storage position,
        TrancheType tranche
    ) internal view returns (uint256) {
        if (tranche == TrancheType.Junior) {
            return position.juniorDeposit;
        } else if (tranche == TrancheType.Mezzanine) {
            return position.mezzanineDeposit;
        } else {
            return position.seniorDeposit;
        }
    }

    function calculateUserYield(address user) external view returns (uint256) {
        UserPosition storage position = userPositions[user];
        
        uint256 timeElapsed = block.timestamp - position.lastYieldCalculation;
        if (timeElapsed == 0) return position.accruedYield;

        uint256 juniorYield = _calculateTrancheYield(position.juniorDeposit, TrancheType.Junior, timeElapsed);
        uint256 mezzanineYield = _calculateTrancheYield(position.mezzanineDeposit, TrancheType.Mezzanine, timeElapsed);
        uint256 seniorYield = _calculateTrancheYield(position.seniorDeposit, TrancheType.Senior, timeElapsed);

        return position.accruedYield + juniorYield + mezzanineYield + seniorYield;
    }

    function _calculateTrancheYield(
        uint256 depositAmount,
        TrancheType tranche,
        uint256 timeElapsed
    ) internal view returns (uint256) {
        if (depositAmount == 0) return 0;
        
        TrancheInfo storage info = tranches[tranche];
        return (depositAmount * info.yieldRate * timeElapsed) / (365 days * 10000);
    }

    function getTrancheAPY(TrancheType tranche) external view returns (uint256) {
        return tranches[tranche].yieldRate;
    }

    function getTrancheRisk(TrancheType tranche) external view returns (uint256) {
        return tranches[tranche].riskMultiplier;
    }

    function getUserPosition(address user) external view returns (
        uint256 juniorDeposit,
        uint256 mezzanineDeposit,
        uint256 seniorDeposit,
        uint256 totalYield,
        uint256 lastDeposit
    ) {
        UserPosition storage position = userPositions[user];
        
        return (
            position.juniorDeposit,
            position.mezzanineDeposit,
            position.seniorDeposit,
            this.calculateUserYield(user),
            position.lastDepositTime
        );
    }

    function getVaultMetrics() external view returns (
        uint256 totalAssets,
        uint256 juniorTVL,
        uint256 mezzanineTVL,
        uint256 seniorTVL,
        uint256 insuranceBalance,
        uint256 totalSlashingEvents
    ) {
        return (
            totalVaultAssets,
            tranches[TrancheType.Junior].currentBalance,
            tranches[TrancheType.Mezzanine].currentBalance,
            tranches[TrancheType.Senior].currentBalance,
            insuranceFund,
            slashingEventCounter
        );
    }

    function getSlashingEvent(uint256 eventId) external view returns (SlashingEvent memory) {
        return slashingEvents[eventId];
    }

    function setYieldRates(
        uint256 juniorRate,
        uint256 mezzanineRate,
        uint256 seniorRate
    ) external onlyOwner {
        tranches[TrancheType.Junior].yieldRate = juniorRate;
        tranches[TrancheType.Mezzanine].yieldRate = mezzanineRate;
        tranches[TrancheType.Senior].yieldRate = seniorRate;
    }

    function setPerformanceFeeRate(uint256 newRate) external onlyOwner {
        require(newRate <= 2000, "Fee cannot exceed 20%");
        performanceFeeRate = newRate;
    }

    function withdrawPerformanceFees() external onlyOwner {
        uint256 amount = insuranceFund;
        insuranceFund = 0;

        depositToken.safeTransfer(owner(), amount);
        emit PerformanceFeeCollected(amount, owner());
    }

    function emergencyPause() external onlyOwner {
        _pause();
        
        tranches[TrancheType.Junior].isActive = false;
        tranches[TrancheType.Mezzanine].isActive = false;
        tranches[TrancheType.Senior].isActive = false;
    }

    function emergencyUnpause() external onlyOwner {
        _unpause();
        
        tranches[TrancheType.Junior].isActive = true;
        tranches[TrancheType.Mezzanine].isActive = true;
        tranches[TrancheType.Senior].isActive = true;
    }

    function getTrancheUtilization(TrancheType tranche) external view returns (uint256) {
        TrancheInfo storage info = tranches[tranche];
        
        if (totalVaultAssets == 0) return 0;
        return (info.currentBalance * 10000) / totalVaultAssets;
    }
}