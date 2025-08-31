// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "fhevm/lib/TFHE.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract GrantPool is ReentrancyGuard, Ownable, Pausable {
    using SafeERC20 for IERC20;

    enum GrantPhase {
        Setup,
        BiddingOpen,
        BiddingClosed,
        WinnerSelected,
        Completed,
        Cancelled
    }

    struct Grant {
        uint256 id;
        string title;
        string description;
        uint256 totalPool;
        address poolToken;
        uint256 maxWinners;
        uint256 biddingStart;
        uint256 biddingEnd;
        uint256 revealEnd;
        GrantPhase phase;
        address[] winners;
        mapping(address => bool) hasSubmitted;
        mapping(address => euint256) encryptedBids;
        mapping(address => bool) isRevealed;
        mapping(address => uint256) revealedBids;
        uint256 submissionCount;
    }

    struct BidSubmission {
        address bidder;
        euint256 encryptedAmount;
        string proposalURI;
        uint256 timestamp;
        bool isRevealed;
        uint256 revealedAmount;
    }

    mapping(uint256 => Grant) public grants;
    mapping(uint256 => mapping(address => BidSubmission)) public bidSubmissions;
    mapping(uint256 => address[]) public grantBidders;
    
    uint256 private _grantCounter;
    uint256 public revealDelay = 24 hours;

    event GrantCreated(
        uint256 indexed grantId,
        string title,
        uint256 totalPool,
        address poolToken,
        uint256 maxWinners,
        uint256 biddingStart,
        uint256 biddingEnd
    );

    event BidSubmitted(
        uint256 indexed grantId,
        address indexed bidder,
        string proposalURI
    );

    event BiddingPhaseEnded(uint256 indexed grantId);

    event BidRevealed(
        uint256 indexed grantId,
        address indexed bidder,
        uint256 amount
    );

    event WinnersSelected(
        uint256 indexed grantId,
        address[] winners,
        uint256[] amounts
    );

    event GrantCompleted(uint256 indexed grantId);

    event GrantCancelled(uint256 indexed grantId);

    error InvalidGrantId();
    error BiddingNotOpen();
    error BiddingStillOpen();
    error AlreadySubmitted();
    error NotSubmitted();
    error RevealPeriodEnded();
    error WinnersAlreadySelected();
    error GrantNotCompleted();
    error InsufficientPool();
    error InvalidTimeParameters();

    constructor() Ownable(msg.sender) {}

    function createGrant(
        string calldata title,
        string calldata description,
        uint256 totalPool,
        address poolToken,
        uint256 maxWinners,
        uint256 biddingDuration,
        uint256 revealDuration
    ) external payable onlyOwner returns (uint256) {
        if (biddingDuration == 0 || revealDuration == 0) {
            revert InvalidTimeParameters();
        }
        if (maxWinners == 0) {
            revert InvalidTimeParameters();
        }

        _grantCounter++;
        uint256 grantId = _grantCounter;

        Grant storage grant = grants[grantId];
        grant.id = grantId;
        grant.title = title;
        grant.description = description;
        grant.totalPool = totalPool;
        grant.poolToken = poolToken;
        grant.maxWinners = maxWinners;
        grant.biddingStart = block.timestamp;
        grant.biddingEnd = block.timestamp + biddingDuration;
        grant.revealEnd = grant.biddingEnd + revealDuration;
        grant.phase = GrantPhase.Setup;

        if (poolToken == address(0)) {
            if (msg.value != totalPool) {
                revert InsufficientPool();
            }
        } else {
            require(msg.value == 0, "ETH sent with token grant");
            IERC20(poolToken).safeTransferFrom(msg.sender, address(this), totalPool);
        }

        grant.phase = GrantPhase.BiddingOpen;

        emit GrantCreated(
            grantId,
            title,
            totalPool,
            poolToken,
            maxWinners,
            grant.biddingStart,
            grant.biddingEnd
        );

        return grantId;
    }

    function submitBid(
        uint256 grantId,
        bytes calldata encryptedAmount,
        string calldata proposalURI
    ) external nonReentrant whenNotPaused {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }
        if (grant.phase != GrantPhase.BiddingOpen) {
            revert BiddingNotOpen();
        }
        if (block.timestamp >= grant.biddingEnd) {
            revert BiddingNotOpen();
        }
        if (grant.hasSubmitted[msg.sender]) {
            revert AlreadySubmitted();
        }

        euint256 bidAmount = TFHE.asEuint256(encryptedAmount);
        
        grant.encryptedBids[msg.sender] = bidAmount;
        grant.hasSubmitted[msg.sender] = true;
        grant.submissionCount++;

        bidSubmissions[grantId][msg.sender] = BidSubmission({
            bidder: msg.sender,
            encryptedAmount: bidAmount,
            proposalURI: proposalURI,
            timestamp: block.timestamp,
            isRevealed: false,
            revealedAmount: 0
        });

        grantBidders[grantId].push(msg.sender);

        emit BidSubmitted(grantId, msg.sender, proposalURI);
    }

    function endBiddingPhase(uint256 grantId) external onlyOwner {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }
        if (grant.phase != GrantPhase.BiddingOpen) {
            revert BiddingNotOpen();
        }
        if (block.timestamp < grant.biddingEnd) {
            revert BiddingStillOpen();
        }

        grant.phase = GrantPhase.BiddingClosed;
        emit BiddingPhaseEnded(grantId);
    }

    function revealBid(uint256 grantId) external {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }
        if (grant.phase != GrantPhase.BiddingClosed) {
            revert BiddingStillOpen();
        }
        if (!grant.hasSubmitted[msg.sender]) {
            revert NotSubmitted();
        }
        if (block.timestamp > grant.revealEnd) {
            revert RevealPeriodEnded();
        }

        BidSubmission storage submission = bidSubmissions[grantId][msg.sender];
        if (submission.isRevealed) {
            return; // Already revealed
        }

        uint256 revealedAmount = TFHE.decrypt(submission.encryptedAmount);
        submission.revealedAmount = revealedAmount;
        submission.isRevealed = true;

        grant.isRevealed[msg.sender] = true;
        grant.revealedBids[msg.sender] = revealedAmount;

        emit BidRevealed(grantId, msg.sender, revealedAmount);
    }

    function selectWinners(uint256 grantId) external onlyOwner {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }
        if (grant.phase != GrantPhase.BiddingClosed) {
            revert BiddingStillOpen();
        }
        if (block.timestamp <= grant.revealEnd) {
            revert RevealPeriodEnded();
        }
        if (grant.winners.length > 0) {
            revert WinnersAlreadySelected();
        }

        address[] memory bidders = grantBidders[grantId];
        uint256[] memory amounts = new uint256[](bidders.length);
        address[] memory validBidders = new address[](bidders.length);
        uint256 validCount = 0;

        for (uint256 i = 0; i < bidders.length; i++) {
            if (grant.isRevealed[bidders[i]]) {
                validBidders[validCount] = bidders[i];
                amounts[validCount] = grant.revealedBids[bidders[i]];
                validCount++;
            }
        }

        _sortBidsByAmount(validBidders, amounts, validCount);

        uint256 winnersToSelect = validCount < grant.maxWinners ? validCount : grant.maxWinners;
        uint256[] memory winnerAmounts = new uint256[](winnersToSelect);

        for (uint256 i = 0; i < winnersToSelect; i++) {
            grant.winners.push(validBidders[i]);
            winnerAmounts[i] = amounts[i];
        }

        grant.phase = GrantPhase.WinnerSelected;

        emit WinnersSelected(grantId, grant.winners, winnerAmounts);
    }

    function distributeGrant(uint256 grantId) external onlyOwner nonReentrant {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }
        if (grant.phase != GrantPhase.WinnerSelected) {
            revert WinnersAlreadySelected();
        }

        uint256 amountPerWinner = grant.totalPool / grant.winners.length;

        for (uint256 i = 0; i < grant.winners.length; i++) {
            address winner = grant.winners[i];
            
            if (grant.poolToken == address(0)) {
                (bool success, ) = winner.call{value: amountPerWinner}("");
                require(success, "ETH transfer failed");
            } else {
                IERC20(grant.poolToken).safeTransfer(winner, amountPerWinner);
            }
        }

        grant.phase = GrantPhase.Completed;
        emit GrantCompleted(grantId);
    }

    function _sortBidsByAmount(
        address[] memory bidders,
        uint256[] memory amounts,
        uint256 length
    ) internal pure {
        for (uint256 i = 0; i < length - 1; i++) {
            for (uint256 j = 0; j < length - i - 1; j++) {
                if (amounts[j] < amounts[j + 1]) {
                    (amounts[j], amounts[j + 1]) = (amounts[j + 1], amounts[j]);
                    (bidders[j], bidders[j + 1]) = (bidders[j + 1], bidders[j]);
                }
            }
        }
    }

    function getGrantDetails(uint256 grantId) external view returns (
        string memory title,
        string memory description,
        uint256 totalPool,
        address poolToken,
        uint256 maxWinners,
        uint256 biddingStart,
        uint256 biddingEnd,
        uint256 revealEnd,
        GrantPhase phase,
        uint256 submissionCount
    ) {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }

        return (
            grant.title,
            grant.description,
            grant.totalPool,
            grant.poolToken,
            grant.maxWinners,
            grant.biddingStart,
            grant.biddingEnd,
            grant.revealEnd,
            grant.phase,
            grant.submissionCount
        );
    }

    function getGrantWinners(uint256 grantId) external view returns (address[] memory) {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }

        return grant.winners;
    }

    function getGrantBidders(uint256 grantId) external view returns (address[] memory) {
        return grantBidders[grantId];
    }

    function hasBidSubmitted(uint256 grantId, address bidder) external view returns (bool) {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }

        return grant.hasSubmitted[bidder];
    }

    function isBidRevealed(uint256 grantId, address bidder) external view returns (bool) {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }

        return grant.isRevealed[bidder];
    }

    function getRevealedBid(uint256 grantId, address bidder) external view returns (uint256) {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }
        if (!grant.isRevealed[bidder]) {
            return 0;
        }

        return grant.revealedBids[bidder];
    }

    function emergencyCancel(uint256 grantId) external onlyOwner {
        Grant storage grant = grants[grantId];
        
        if (grant.id == 0) {
            revert InvalidGrantId();
        }
        if (grant.phase == GrantPhase.Completed || grant.phase == GrantPhase.Cancelled) {
            revert GrantNotCompleted();
        }

        grant.phase = GrantPhase.Cancelled;

        if (grant.poolToken == address(0)) {
            (bool success, ) = owner().call{value: grant.totalPool}("");
            require(success, "ETH refund failed");
        } else {
            IERC20(grant.poolToken).safeTransfer(owner(), grant.totalPool);
        }

        emit GrantCancelled(grantId);
    }

    function setRevealDelay(uint256 newDelay) external onlyOwner {
        revealDelay = newDelay;
    }

    function getGrantCount() external view returns (uint256) {
        return _grantCounter;
    }
}