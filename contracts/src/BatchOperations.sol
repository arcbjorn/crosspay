// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./ConfidentialPayments.sol";
import "./RelayValidator.sol";
import "./TrancheVault.sol";
import "./PaymentCore.sol";
import "fhevm/lib/TFHE.sol";

contract BatchOperations {
    ConfidentialPayments public immutable confidentialPayments;
    RelayValidator public immutable relayValidator;
    TrancheVault public immutable trancheVault;
    PaymentCore public immutable paymentCore;

    uint256 private constant FEE_BPS = 10; // Match PaymentCore fee

    struct BatchPayment {
        address recipient;
        address token;
        einput encryptedAmount;
        bytes inputProof;
        string metadataURI;
        bool makePrivate;
    }

    struct BatchValidation {
        uint256 paymentId;
        bytes32 messageHash;
        uint256 amount;
    }

    struct BatchDeposit {
        TrancheVault.TrancheType tranche;
        uint256 amount;
    }

    event BatchPaymentsCreated(uint256[] paymentIds, address indexed sender);
    event BatchValidationsRequested(uint256[] requestIds, address indexed requester);
    event BatchDepositsCompleted(address indexed user, uint256 totalAmount);

    error BatchSizeMismatch();
    error BatchTooLarge();
    error BatchExecutionFailed();

    constructor(
        address _confidentialPayments,
        address _relayValidator,
        address _trancheVault,
        address _paymentCore
    ) {
        confidentialPayments = ConfidentialPayments(_confidentialPayments);
        relayValidator = RelayValidator(_relayValidator);
        trancheVault = TrancheVault(_trancheVault);
        paymentCore = PaymentCore(_paymentCore);
    }

    function batchCreatePayments(
        BatchPayment[] calldata payments
    ) external payable returns (uint256[] memory paymentIds) {
        if (payments.length == 0 || payments.length > 50) {
            revert BatchTooLarge();
        }

        paymentIds = new uint256[](payments.length);

        // Count ETH payments to split msg.value
        uint256 ethCount = 0;
        for (uint256 i = 0; i < payments.length; i++) {
            if (payments[i].token == address(0)) ethCount++;
        }
        uint256 perShare = ethCount > 0 ? (msg.value / ethCount) : 0;

        for (uint256 i = 0; i < payments.length; i++) {
            BatchPayment memory p = payments[i];

            // If no FHE proof provided and ETH payment, create public payment via PaymentCore
            if (p.token == address(0) && p.inputProof.length == 0 && !p.makePrivate) {
                // Compute amount so that amount + fee == perShare
                // amount = perShare * 10000 / (10000 + FEE_BPS)
                uint256 amount = perShare * 10000 / (10000 + FEE_BPS);
                uint256 fee = (amount * FEE_BPS) / 10000;
                uint256 total = amount + fee;
                try paymentCore.createPayment{value: total}(
                    p.recipient,
                    address(0),
                    amount,
                    p.metadataURI,
                    "",
                    ""
                ) returns (uint256 pid) {
                    paymentIds[i] = pid;
                } catch {
                    revert BatchExecutionFailed();
                }
                continue;
            }

            // Otherwise, call confidential route (requires valid FHE input)
            try confidentialPayments.createConfidentialPayment{
                value: 0
            }(
                p.recipient,
                p.token,
                p.encryptedAmount,
                p.inputProof,
                p.metadataURI,
                p.makePrivate
            ) returns (uint256 pid) {
                paymentIds[i] = pid;
            } catch {
                revert BatchExecutionFailed();
            }
        }

        emit BatchPaymentsCreated(paymentIds, msg.sender);
    }

    function batchRequestValidations(
        BatchValidation[] calldata validations
    ) external returns (uint256[] memory requestIds) {
        if (validations.length == 0 || validations.length > 20) {
            revert BatchTooLarge();
        }

        // Only owner of RelayValidator can request validations
        require(msg.sender == relayValidator.owner(), "Only validator owner");

        requestIds = new uint256[](validations.length);

        for (uint256 i = 0; i < validations.length; i++) {
            BatchValidation memory validation = validations[i];
            
            try relayValidator.requestValidation(
                validation.paymentId,
                validation.messageHash,
                validation.amount
            ) returns (uint256 requestId) {
                requestIds[i] = requestId;
            } catch {
                revert BatchExecutionFailed();
            }
        }

        emit BatchValidationsRequested(requestIds, msg.sender);
    }

    function batchDeposit(
        BatchDeposit[] calldata deposits
    ) external {
        if (deposits.length == 0 || deposits.length > 10) {
            revert BatchTooLarge();
        }

        uint256 totalAmount = 0;

        for (uint256 i = 0; i < deposits.length; i++) {
            BatchDeposit memory d = deposits[i];
            totalAmount += d.amount;

            try trancheVault.depositFor(msg.sender, d.tranche, d.amount) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }

        emit BatchDepositsCompleted(msg.sender, totalAmount);
    }

    function batchGrantDisclosurePermissions(
        uint256[] calldata paymentIds,
        address[] calldata viewers
    ) external {
        if (paymentIds.length != viewers.length) {
            revert BatchSizeMismatch();
        }
        if (paymentIds.length > 100) {
            revert BatchTooLarge();
        }

        for (uint256 i = 0; i < paymentIds.length; i++) {
            try confidentialPayments.grantDisclosurePermission(viewers[i], paymentIds[i]) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }
    }

    function batchRevokeDisclosurePermissions(
        uint256[] calldata paymentIds,
        address[] calldata viewers
    ) external {
        if (paymentIds.length != viewers.length) {
            revert BatchSizeMismatch();
        }
        if (paymentIds.length > 100) {
            revert BatchTooLarge();
        }

        for (uint256 i = 0; i < paymentIds.length; i++) {
            try confidentialPayments.revokeDisclosurePermission(viewers[i], paymentIds[i]) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }
    }

    function batchCompletePayments(uint256[] calldata paymentIds) external {
        if (paymentIds.length > 50) {
            revert BatchTooLarge();
        }

        for (uint256 i = 0; i < paymentIds.length; i++) {
            try confidentialPayments.completeConfidentialPayment(paymentIds[i]) {
                // Success
            } catch {
                revert BatchExecutionFailed();
            }
        }
    }

    function estimateBatchGas(
        BatchPayment[] calldata payments
    ) external pure returns (uint256 totalGas) {
        // Rough estimation: 180k gas per payment + 21k base + batch overhead
        totalGas = 21000 + (payments.length * 185000) + 5000;
    }
}
